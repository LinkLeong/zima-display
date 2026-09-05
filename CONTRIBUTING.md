# Contributing

Contributions are welcome. Before opening a pull request, please search existing issues and keep changes focused on one problem.

## Development checks

```bash
gofmt -w ./cmd ./internal
go test ./...
node --check web/app.js
node scripts/check-i18n.mjs
node scripts/test-web-language.mjs
sh -n scripts/*.sh packaging/raw/usr/libexec/zima-display/start-renderer
```

Build the native package when changing packaging, renderer or frontend files:

```bash
./scripts/build.sh amd64
```

## Pull requests

- Explain the user-visible behavior and the ZimaOS environment used for testing.
- Add or update tests for configuration, API and renderer logic.
- Update both README files when changing installation or usage.
- Update both locale dictionaries when adding visible web text.
- Never commit server passwords, private IP details, CasaOS tokens, logs or uploaded media.

Use conventional, concise commit subjects such as `feat: add dashboard language setting` or `fix: reconnect stale mpv socket`.
