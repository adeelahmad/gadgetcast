# Contributing to gadgetcast

Thanks for helping. The most useful contributions right now are hardware reports, bug reports with logs, and small focused PRs.

## Ways to help

- **Hardware reports.** Tell us which Pi, which host OS and which app (Zoom, Teams, OBS, a browser) worked or didn't. Use the [hardware report](https://github.com/adeelahmad/gadgetcast/issues/new?template=hardware-report.yml) template.
- **Bugs.** Open a [bug report](https://github.com/adeelahmad/gadgetcast/issues/new?template=bug.yml) with steps, what you expected and the relevant `journalctl -u gadgetcast` output.
- **Ideas.** Open a [feature request](https://github.com/adeelahmad/gadgetcast/issues/new?template=feature.yml). For bigger changes, please open an issue before writing code so we can agree on the shape.

## Development

You need Go 1.25 or newer. Node 22 is needed only if you work on the web UI in `web/`.

```console
git clone https://github.com/adeelahmad/gadgetcast && cd gadgetcast
gofmt -l .            # must print nothing
go vet ./...
go test -race -cover ./...
go run ./cmd/gadgetcast -dev -data ./data
```

For the UI: `cd web && npm ci && npm run lint && npm test && npm run build`.

The design, constraints and roadmap live in [docs/design.md](docs/design.md). Code under `internal/` is pure Go with no hardware dependencies, so it tests on any machine.

## Pull requests

- Keep each PR to one change, with tests. New logic is written test-first and pure-logic packages stay above 80% coverage.
- Wrap errors with context (`fmt.Errorf("…: %w", err)`), never swallow them, and validate input at the HTTP and WebSocket boundary.
- Don't add third-party dependencies without discussing it in an issue first.
- CI (gofmt, vet, build, race tests, and the web checks) must be green.

## Security issues

Please don't open public issues for vulnerabilities. See [SECURITY.md](SECURITY.md).

By contributing you agree that your contributions are licensed under the [MIT License](LICENSE).
