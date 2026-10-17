#!/bin/sh
# gadgetcast installer and updater.
#
#   curl -fsSL https://raw.githubusercontent.com/adeelahmad/gadgetcast/master/install.sh | sh
#
# One command installs the latest release, enables USB gadget mode on a
# Raspberry Pi, starts the service and turns on automatic updates. Running it
# again updates in place. No prompts.
#
# Options (flags or environment variables):
#   --update                 only update the binary if a newer release exists (used by the timer)
#   --uninstall              remove gadgetcast, its service and the update timer
#   --reboot                 reboot automatically if USB gadget mode was just enabled
#   --init                   create config, TLS cert and pairing PIN if missing (run by the service)
#   --boot-config            apply gadgetcast.env from the boot partition (run at boot, see below)
#   GADGETCAST_VERSION=v1.2.3  install a specific release instead of the latest
#   GADGETCAST_NO_AUTOUPDATE=1 do not install the daily update timer
#   GADGETCAST_BINARY=/path    install this local binary instead of downloading (image builds)
#   GADGETCAST_PI=1            force Raspberry Pi boot configuration (image builds / chroot)
#   GADGETCAST_ARCH=arm64      override detected architecture
#   GADGETCAST_REPO=owner/name install from a fork
#
# Headless setup without Raspberry Pi Imager: put a gadgetcast.env file on the
# SD card's boot partition. It is applied early on the next boot and then
# deleted, because it may hold passwords. KEY=value lines, all optional:
#   HOSTNAME=studio  WIFI_SSID=MyNetwork  WIFI_PASSWORD=secret  WIFI_COUNTRY=GB
#   PIN=123456       GADGETCAST_ARGS="-addr :443"
set -eu

REPO="${GADGETCAST_REPO:-adeelahmad/gadgetcast}"
BASE_URL="${GADGETCAST_BASE_URL:-https://github.com/$REPO/releases}"
SCRIPT_URL="https://raw.githubusercontent.com/$REPO/master/install.sh"

BIN=/usr/local/bin/gadgetcast
UPDATER=/usr/local/sbin/gadgetcast-update
SHARE=/usr/local/share/gadgetcast
VERSION_FILE="$SHARE/VERSION"
ENV_FILE=/etc/default/gadgetcast
CONF_DIR=/etc/gadgetcast
DATA_DIR=/var/lib/gadgetcast
UNIT_DIR=/etc/systemd/system

MODE=install
REBOOT="${GADGETCAST_REBOOT:-0}"
for arg in "$@"; do
	case "$arg" in
	--update) MODE=update ;;
	--uninstall) MODE=uninstall ;;
	--reboot) REBOOT=1 ;;
	--init) MODE=init ;;
	--boot-config) MODE=boot-config ;;
	-h | --help)
		sed -n '2,27p' "$0" 2>/dev/null | sed 's/^# \{0,1\}//'
		exit 0
		;;
	*)
		echo "gadgetcast: unknown option: $arg" >&2
		exit 2
		;;
	esac
done

say() { [ "$MODE" = update ] || printf '\033[1m==>\033[0m %s\n' "$*"; }
die() {
	printf 'gadgetcast: %s\n' "$*" >&2
	exit 1
}
have() { command -v "$1" >/dev/null 2>&1; }

# Re-run as root so the user never has to type sudo. When piped from curl there
# is no script file to re-exec, so fetch it again.
if [ "$(id -u)" -ne 0 ]; then
	have sudo || die "please run as root"
	if [ -f "$0" ] && [ "$(basename "$0")" != sh ]; then
		exec sudo -E sh "$0" "$@"
	fi
	tmp="$(mktemp)"
	curl -fsSL "$SCRIPT_URL" -o "$tmp" || die "could not download installer"
	exec sudo -E sh "$tmp" "$@"
fi

systemd_running() { [ -d /run/systemd/system ]; }

uninstall() {
	if have systemctl; then
		systemctl disable --now gadgetcast.service gadgetcast-update.timer gadgetcast-boot-config.service gadgetcast-gadget.service 2>/dev/null || true
	fi
	rm -f "$BIN" "$UPDATER" "$UNIT_DIR/gadgetcast.service" "$UNIT_DIR/gadgetcast-boot-config.service" \
		"$UNIT_DIR/gadgetcast-gadget.service" \
		"$UNIT_DIR/gadgetcast-update.service" "$UNIT_DIR/gadgetcast-update.timer" \
		/etc/modules-load.d/gadgetcast.conf
	rm -rf "$SHARE"
	if systemd_running; then systemctl daemon-reload; fi
	say "gadgetcast removed (config in $CONF_DIR, data in $DATA_DIR and the dwc2 boot overlay were left in place)"
}

detect_arch() {
	if [ -n "${GADGETCAST_ARCH:-}" ]; then
		echo "$GADGETCAST_ARCH"
		return
	fi
	case "$(uname -m)" in
	aarch64 | arm64) echo arm64 ;;
	armv7l | armv8l) echo armv7 ;;
	armv6l) echo armv6 ;;
	x86_64 | amd64) echo amd64 ;;
	*) die "unsupported architecture: $(uname -m)" ;;
	esac
}

# Latest release tag, resolved from the /releases/latest redirect so no API
# token or rate limit is involved. /latest skips prereleases: a fresh install
# falls back to the newest prerelease (from the releases feed), while the
# update timer prints nothing so it only ever follows full releases.
latest_version() {
	url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$BASE_URL/latest" 2>/dev/null)" || url=
	tag="${url##*/}"
	case "$tag" in
	v*)
		echo "$tag"
		return
		;;
	esac
	[ "$MODE" = update ] && return
	tag="$(curl -fsSL "$BASE_URL.atom" 2>/dev/null | grep -o '/releases/tag/v[^"<]*' | head -n 1)" ||
		die "could not reach GitHub"
	tag="${tag##*/}"
	[ -n "$tag" ] || die "no published release found at $BASE_URL"
	say "no full release yet; installing prerelease $tag" >&2
	echo "$tag"
}

installed_version() { cat "$VERSION_FILE" 2>/dev/null || true; }

sha256() {
	if have sha256sum; then
		sha256sum "$1" | cut -d' ' -f1
	else
		shasum -a 256 "$1" | cut -d' ' -f1
	fi
}

# Downloads and verifies the release tarball, then swaps the binary and the
# updater atomically.
fetch_release() {
	version="$1"
	arch="$2"
	asset="gadgetcast_linux_${arch}.tar.gz"
	work="$(mktemp -d)"
	trap 'rm -rf "$work"' EXIT
	curl -fsSL "$BASE_URL/download/$version/$asset" -o "$work/$asset" ||
		die "download failed: $asset ($version)"
	curl -fsSL "$BASE_URL/download/$version/SHA256SUMS" -o "$work/SHA256SUMS" ||
		die "download failed: SHA256SUMS ($version)"
	want="$(grep " $asset\$" "$work/SHA256SUMS" | cut -d' ' -f1)"
	[ -n "$want" ] || die "$asset missing from SHA256SUMS"
	[ "$(sha256 "$work/$asset")" = "$want" ] || die "checksum mismatch for $asset"
	tar -xzf "$work/$asset" -C "$work"
	[ -f "$work/gadgetcast" ] || die "release archive has no gadgetcast binary"
	install_file "$work/gadgetcast" "$BIN" 755
	if [ -f "$work/install.sh" ]; then
		install_file "$work/install.sh" "$UPDATER" 755
	fi
}

install_file() {
	mkdir -p "$(dirname "$2")"
	cp "$1" "$2.new"
	chmod "$3" "$2.new"
	mv -f "$2.new" "$2"
}

is_pi() {
	[ "${GADGETCAST_PI:-0}" = 1 ] && return 0
	grep -q "Raspberry Pi" /proc/device-tree/model 2>/dev/null
}

# Enables the dwc2 controller in peripheral mode and loads libcomposite so the
# Pi can present itself as a USB webcam + audio device. Prints whether a reboot
# is needed.
configure_pi() {
	needs_reboot=0
	cfg=/boot/firmware/config.txt
	[ -f "$cfg" ] || cfg=/boot/config.txt
	# Match our own line: stock Raspberry Pi OS already has a [cm5]
	# dtoverlay=dwc2,dr_mode=host, which must not count as gadget mode.
	if [ -f "$cfg" ] && ! grep -q '^dtoverlay=dwc2,dr_mode=peripheral' "$cfg"; then
		printf '\n[all]\n# gadgetcast: USB gadget (webcam + audio) mode\ndtoverlay=dwc2,dr_mode=peripheral\n' >>"$cfg"
		needs_reboot=1
	fi
	mkdir -p /etc/modules-load.d
	if [ ! -f /etc/modules-load.d/gadgetcast.conf ]; then
		printf 'dwc2\nlibcomposite\n' >/etc/modules-load.d/gadgetcast.conf
		needs_reboot=1
	fi
	echo "$needs_reboot"
}

boot_dir() {
	for d in /boot/firmware /boot; do
		if [ -f "$d/config.txt" ]; then
			echo "$d"
			return
		fi
	done
}

# First-run setup, idempotent: a self-signed TLS cert for <hostname>.local, a
# random pairing PIN and a config that passes config.Validate. Runs on the
# device itself (service ExecStartPre) so every flashed card gets its own key.
init_config() {
	mkdir -p "$CONF_DIR" "$DATA_DIR"
	chmod 700 "$CONF_DIR"
	if [ ! -s "$CONF_DIR/pin" ]; then
		pin="$(od -An -N4 -tu4 /dev/urandom | tr -d ' ')"
		printf '%06d\n' "$((pin % 1000000))" >"$CONF_DIR/pin"
		chmod 600 "$CONF_DIR/pin"
	fi
	# Reissue the cert when the hostname changes, e.g. one set in Raspberry Pi
	# Imager's customisation screen and applied by cloud-init on first boot.
	host="$(hostname)"
	if [ ! -s "$CONF_DIR/tls.key" ] || [ ! -s "$CONF_DIR/tls.crt" ] ||
		[ "$(cat "$CONF_DIR/tls.host" 2>/dev/null)" != "$host" ]; then
		have openssl || die "openssl is required to create the TLS certificate"
		openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes -days 3650 \
			-subj "/CN=$host.local" \
			-addext "subjectAltName=DNS:$host.local,DNS:$host,DNS:localhost,IP:127.0.0.1" \
			-keyout "$CONF_DIR/tls.key" -out "$CONF_DIR/tls.crt" >/dev/null 2>&1 ||
			die "could not create TLS certificate"
		chmod 600 "$CONF_DIR/tls.key"
		echo "$host" >"$CONF_DIR/tls.host"
	fi
	if [ ! -s "$CONF_DIR/config.json" ]; then
		cat >"$CONF_DIR/config.json" <<EOF
{
  "https": {"enabled": true, "cert_path": "$CONF_DIR/tls.crt", "key_path": "$CONF_DIR/tls.key"},
  "bind": {"mode": "all"},
  "pin": {"required": true, "source": "file:$CONF_DIR/pin"}
}
EOF
	fi
	# Put the PIN where a headless user can read it: the SD card's boot partition.
	bd="$(boot_dir)"
	if [ -n "$bd" ]; then
		note="$(printf 'gadgetcast pairing PIN: %s\nOpen https://%s.local in a browser.' \
			"$(cat "$CONF_DIR/pin")" "$host")"
		[ "$(cat "$bd/gadgetcast-pin.txt" 2>/dev/null)" = "$note" ] ||
			echo "$note" >"$bd/gadgetcast-pin.txt" 2>/dev/null || true
	fi
}

# Value of KEY in a KEY=value env file, with surrounding quotes removed. The
# file is parsed, never sourced, so it cannot run commands.
env_value() {
	sed -n "s/^[[:space:]]*\(export[[:space:]]\{1,\}\)\{0,1\}$1[[:space:]]*=[[:space:]]*//p" "$2" |
		tail -n 1 | tr -d '\r' | sed -e 's/^"\(.*\)"$/\1/' -e "s/^'\(.*\)'\$/\1/"
}

# Applies <boot>/gadgetcast.env (see the header) and deletes it. Runs before
# networking starts, so Wi-Fi and the hostname are in place for this boot.
boot_config() {
	bd="$(boot_dir)"
	f="$bd/gadgetcast.env"
	[ -n "$bd" ] && [ -f "$f" ] || return 0

	host="$(env_value HOSTNAME "$f")"
	if [ -n "$host" ]; then
		case "$host" in
		*[!A-Za-z0-9-]* | -* | *-) die "gadgetcast.env: invalid HOSTNAME: $host" ;;
		esac
		echo "$host" >/etc/hostname
		hostname "$host" 2>/dev/null || true
		if grep -q '^127\.0\.1\.1' /etc/hosts 2>/dev/null; then
			sed -i "s/^127\.0\.1\.1.*/127.0.1.1\t$host/" /etc/hosts
		else
			printf '127.0.1.1\t%s\n' "$host" >>/etc/hosts
		fi
	fi

	country="$(env_value WIFI_COUNTRY "$f")"
	if [ -n "$country" ]; then
		if have raspi-config; then
			raspi-config nonint do_wifi_country "$country" >/dev/null 2>&1 || true
		elif have iw; then
			iw reg set "$country" 2>/dev/null || true
		fi
	fi

	ssid="$(env_value WIFI_SSID "$f")"
	if [ -n "$ssid" ]; then
		psk="$(env_value WIFI_PASSWORD "$f")"
		nm=/etc/NetworkManager/system-connections/gadgetcast-wifi.nmconnection
		mkdir -p "$(dirname "$nm")"
		{
			printf '[connection]\nid=%s\ntype=wifi\nautoconnect=true\n\n' "$ssid"
			printf '[wifi]\nmode=infrastructure\nssid=%s\n\n' "$ssid"
			if [ -n "$psk" ]; then
				printf '[wifi-security]\nkey-mgmt=wpa-psk\npsk=%s\n\n' "$psk"
			fi
			printf '[ipv4]\nmethod=auto\n\n[ipv6]\nmethod=auto\n'
		} >"$nm"
		chmod 600 "$nm"
	fi

	pin="$(env_value PIN "$f")"
	if [ -n "$pin" ]; then
		case "$pin" in
		*[!0-9]*) die "gadgetcast.env: PIN must be digits" ;;
		esac
		mkdir -p "$CONF_DIR"
		chmod 700 "$CONF_DIR"
		echo "$pin" >"$CONF_DIR/pin"
		chmod 600 "$CONF_DIR/pin"
	fi

	if grep -Eq '^[[:space:]]*(export[[:space:]]+)?GADGETCAST_ARGS[[:space:]]*=' "$f"; then
		mkdir -p "$(dirname "$ENV_FILE")"
		printf '# Extra command-line arguments for gadgetcast\nGADGETCAST_ARGS="%s"\n' \
			"$(env_value GADGETCAST_ARGS "$f")" >"$ENV_FILE"
	fi

	rm -f "$f"
	echo "gadgetcast: applied gadgetcast.env from $bd"
}

# Listen on 443 so the URL needs no port, unless something already uses it.
default_args() {
	if have ss && ss -ltnH 2>/dev/null | awk '{print $4}' | grep -q ':443$'; then
		echo ""
	else
		echo "-addr :443"
	fi
}

url() {
	port=""
	grep -q -- '-addr :443' "$ENV_FILE" 2>/dev/null || port=":8443"
	echo "https://$(hostname).local$port"
}

ensure_mdns() {
	have avahi-daemon && return 0
	have apt-get || return 0
	say "installing avahi for $(hostname).local discovery"
	DEBIAN_FRONTEND=noninteractive apt-get install -y -qq avahi-daemon >/dev/null 2>&1 || true
}

# GStreamer decodes the on-air browser's Opus audio and H.264 video and plays
# them into the USB microphone and webcam (internal/media).
ensure_media() {
	have gst-launch-1.0 && gst-inspect-1.0 uvcsink >/dev/null 2>&1 && return 0
	have apt-get || return 0
	say "installing GStreamer for USB audio"
	DEBIAN_FRONTEND=noninteractive apt-get install -y -qq gstreamer1.0-tools gstreamer1.0-plugins-base \
		gstreamer1.0-plugins-good gstreamer1.0-plugins-bad gstreamer1.0-libav gstreamer1.0-alsa >/dev/null 2>&1 ||
		say "could not install GStreamer; USB audio and video stay off until it is installed"
}

write_units() {
	mkdir -p "$UNIT_DIR"
	cat >"$UNIT_DIR/gadgetcast.service" <<EOF
[Unit]
Description=gadgetcast WebRTC to USB webcam/audio bridge
Wants=network-online.target
# cloud-init applies Raspberry Pi Imager's hostname, user and Wi-Fi settings.
After=network-online.target cloud-init.target

[Service]
EnvironmentFile=-$ENV_FILE
ExecStartPre=$UPDATER --init
ExecStart=$BIN -config $CONF_DIR/config.json -data $DATA_DIR \$GADGETCAST_ARGS
Restart=on-failure
RestartSec=2

[Install]
WantedBy=multi-user.target
EOF
	[ -f "$ENV_FILE" ] || printf '# Extra command-line arguments for gadgetcast\nGADGETCAST_ARGS="%s"\n' "$(default_args)" >"$ENV_FILE"

	cat >"$UNIT_DIR/gadgetcast-boot-config.service" <<EOF
[Unit]
Description=Apply gadgetcast.env from the boot partition
DefaultDependencies=no
After=local-fs.target
Before=network-pre.target NetworkManager.service cloud-init-local.service gadgetcast.service
Wants=network-pre.target
ConditionPathExists=|/boot/firmware/gadgetcast.env
ConditionPathExists=|/boot/gadgetcast.env

[Service]
Type=oneshot
ExecStart=$UPDATER --boot-config

[Install]
WantedBy=multi-user.target
EOF

	# Creates the USB webcam + mic + speaker the host sees (configfs gadget).
	cat >"$UNIT_DIR/gadgetcast-gadget.service" <<EOF
[Unit]
Description=gadgetcast USB gadget (webcam, microphone, speaker)
DefaultDependencies=no
Wants=sys-kernel-config.mount
After=sys-kernel-config.mount systemd-modules-load.service
Before=gadgetcast.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=$BIN gadget up
ExecStop=$BIN gadget down

[Install]
WantedBy=multi-user.target
EOF

	if [ "${GADGETCAST_NO_AUTOUPDATE:-0}" = 1 ]; then
		rm -f "$UNIT_DIR/gadgetcast-update.service" "$UNIT_DIR/gadgetcast-update.timer"
		return
	fi
	cat >"$UNIT_DIR/gadgetcast-update.service" <<EOF
[Unit]
Description=Update gadgetcast to the latest release
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=$UPDATER --update
EOF
	cat >"$UNIT_DIR/gadgetcast-update.timer" <<EOF
[Unit]
Description=Check for gadgetcast updates daily

[Timer]
OnBootSec=10min
OnUnitActiveSec=1d
RandomizedDelaySec=1h
Persistent=true

[Install]
WantedBy=timers.target
EOF
}

enable_units() {
	have systemctl || return 0
	units="gadgetcast.service gadgetcast-boot-config.service"
	[ -f "$UNIT_DIR/gadgetcast-update.timer" ] && units="$units gadgetcast-update.timer"
	is_pi && units="$units gadgetcast-gadget.service"
	if systemd_running; then
		systemctl daemon-reload
		# shellcheck disable=SC2086
		systemctl enable --now $units >/dev/null 2>&1 ||
			say "service installed but did not start; see: journalctl -u gadgetcast"
	else
		# chroot / image build: enable for next boot only
		# shellcheck disable=SC2086
		systemctl enable $units >/dev/null 2>&1 || true
	fi
}

restart_service() {
	if systemd_running && systemctl is-enabled gadgetcast.service >/dev/null 2>&1; then
		systemctl restart gadgetcast.service || true
	fi
}

main() {
	if [ "$MODE" = uninstall ]; then
		uninstall
		return
	fi
	if [ "$MODE" = init ]; then
		init_config
		return
	fi
	if [ "$MODE" = boot-config ]; then
		boot_config
		return
	fi

	arch="$(detect_arch)"
	current="$(installed_version)"

	if [ -n "${GADGETCAST_BINARY:-}" ]; then
		version="${GADGETCAST_VERSION:-dev}"
		[ "$GADGETCAST_BINARY" = "$BIN" ] || install_file "$GADGETCAST_BINARY" "$BIN" 755
		[ "$0" = "$UPDATER" ] || install_file "$0" "$UPDATER" 755
	else
		have curl || die "curl is required"
		version="${GADGETCAST_VERSION:-$(latest_version)}"
		if [ -z "$version" ]; then
			# latest_version's die only left its subshell; updates without a full release are a no-op
			[ "$MODE" = update ] || die "no release to install"
			echo "gadgetcast: no full release to update to; staying on ${current:-none}"
			return
		fi
		if [ "$version" = "$current" ] && [ -x "$BIN" ]; then
			say "gadgetcast $version is already the latest"
			[ "$MODE" = update ] && return
		else
			say "downloading gadgetcast $version ($arch)"
			fetch_release "$version" "$arch"
			if [ "$MODE" = update ]; then
				mkdir -p "$SHARE"
				echo "$version" >"$VERSION_FILE"
				restart_service
				echo "gadgetcast updated: ${current:-none} -> $version"
				return
			fi
		fi
	fi
	mkdir -p "$SHARE"
	echo "$version" >"$VERSION_FILE"

	needs_reboot=0
	if is_pi; then
		say "enabling USB gadget mode"
		needs_reboot="$(configure_pi)"
	fi
	if [ -z "${GADGETCAST_BINARY:-}" ]; then
		ensure_mdns
		if is_pi; then ensure_media; fi
	fi

	write_units
	enable_units
	restart_service

	say "gadgetcast $version installed"
	[ "${GADGETCAST_NO_AUTOUPDATE:-0}" = 1 ] || say "automatic updates are on (daily)"
	if [ -z "${GADGETCAST_BINARY:-}" ]; then
		init_config
		say "open $(url) in a browser and pair with PIN $(cat "$CONF_DIR/pin")"
	fi
	if [ "$needs_reboot" = 1 ] && [ -z "${GADGETCAST_BINARY:-}" ]; then
		if [ "$REBOOT" = 1 ]; then
			say "rebooting to enable USB gadget mode"
			reboot
		else
			say "reboot once to enable USB gadget mode: sudo reboot"
		fi
	fi
}

main
