# Contributing

Thanks for your interest in contributing.

## Development setup

1. Clone the repository
2. Copy `.env.example` to `.env` and adjust values
3. Install Go 1.25+ and Node.js 20+
4. Run `make install-hooks` to install pre-commit hooks
5. Start the stack: `docker compose up -d`

## Code style

- Run `gofmt -w .` before committing (enforced by pre-commit hook)
- Follow standard Go conventions
- Keep service boundaries clear: domain -> repository -> service -> server

## Testing

- Add unit tests for new business logic in `internal/*/service/`
- Use mock repositories to isolate from infrastructure
- Run `go test -race ./...` before pushing

## Commit messages

Use conventional commits prefix:
- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation only
- `chore:` build process, dependencies
- `refactor:` code change that neither fixes a bug nor adds a feature
- `test:` adding or fixing tests

## Pull Requests

1. Create a feature branch from `main`
2. Make your changes with tests
3. Ensure CI passes (lint, tests, build)
4. Open a PR with a clear description

## Reporting issues

Use GitHub Issues. Include reproduction steps and environment details.
