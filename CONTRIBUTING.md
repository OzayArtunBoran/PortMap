# Contributing to PortMap

Thanks for your interest in contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/ozayartunboran/portmap.git
cd portmap
make build
```

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Ensure linting passes: `make lint`
6. Commit with conventional commits (see below)
7. Open a Pull Request

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation only
- `test:` — Adding or updating tests
- `chore:` — Maintenance (CI, deps, config)
- `refactor:` — Code change that neither fixes a bug nor adds a feature

Examples:

```
feat: add Slack notification support
fix: handle empty config file gracefully
docs: update CLI reference in README
test: add edge case tests for config parser
chore: update CI workflow to Go 1.22
```

## Code Style

- Follow standard Go conventions
- Run `go vet ./...` before committing
- Keep functions small and focused

## Reporting Issues

- Use the [Bug Report](.github/ISSUE_TEMPLATE/bug_report.md) template for bugs
- Use the [Feature Request](.github/ISSUE_TEMPLATE/feature_request.md) template for ideas
- Check existing issues before creating a new one

## License

By contributing, you agree that your contributions will be licensed under the project's [MIT](LICENSE) license.
