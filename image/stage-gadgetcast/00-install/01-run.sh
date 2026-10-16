#!/bin/bash -e
# Bakes the release binary into the image and runs the regular installer in
# offline mode, so a flashed card behaves exactly like `install.sh` on a Pi.
# files/ is filled by .github/workflows/pi-image.yml.

install -D -m 755 files/gadgetcast "${ROOTFS_DIR}/usr/local/bin/gadgetcast"
install -D -m 755 files/install.sh "${ROOTFS_DIR}/usr/local/sbin/gadgetcast-update"

on_chroot <<CHROOT
GADGETCAST_BINARY=/usr/local/bin/gadgetcast \
GADGETCAST_VERSION="$(cat files/VERSION)" \
GADGETCAST_PI=1 \
GADGETCAST_ARCH="$(cat files/ARCH 2>/dev/null || echo arm64)" \
sh /usr/local/sbin/gadgetcast-update
CHROOT

# Headless setup without Imager: rename to gadgetcast.env and fill in.
cat >"${ROOTFS_DIR}/boot/firmware/gadgetcast.env.example" <<'ENV'
# Rename this file to gadgetcast.env to apply it on the next boot. It is
# deleted once applied. Every line is optional.
#HOSTNAME=gadgetcast
#WIFI_SSID=MyNetwork
#WIFI_PASSWORD=secret
#WIFI_COUNTRY=GB
#PIN=123456
#GADGETCAST_ARGS="-addr :443"
ENV
