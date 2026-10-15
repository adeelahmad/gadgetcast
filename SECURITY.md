# Security policy

gadgetcast is, by design, a network-controlled camera and microphone for whatever computer it is plugged into. We treat anything that lets an unpaired client reach the USB outputs, read the host's audio, or bypass pairing as a serious vulnerability.

## Reporting

Please report vulnerabilities privately through [GitHub security advisories](https://github.com/adeelahmad/gadgetcast/security/advisories/new). Do not open a public issue.

Include the version (`gadgetcast -version`), your configuration with secrets removed, and steps to reproduce. We aim to acknowledge reports within a few days and will credit you in the release notes unless you'd rather we didn't.

## Supported versions

Only the latest release receives security fixes. Installs from `install.sh` update themselves daily unless `GADGETCAST_NO_AUTOUPDATE=1` was set.

## Hardening tips

- Keep HTTPS on and PIN pairing required.
- On shared networks, bind the server to your Tailscale interface only.
- Treat the on-air indicator as the source of truth for whether the host is receiving media.
