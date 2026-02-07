# Contributing

Thanks for your interest in contributing!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/errx.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run checks: `make ci`
6. Commit with [conventional commits](https://www.conventionalcommits.org/): `git commit -m "feat: add feature"`
7. Push and open a PR

## Development

```bash
make test      # run tests
make lint      # run linter
make cover     # run tests with coverage
make generate  # run go generate
make ci        # run all checks
```

## Guidelines

- Write tests for new functionality
- Keep changes focused and minimal
- Update documentation if needed
- Follow existing code style

## Reporting Issues

- Search existing issues first
- Include Go version and OS
- Provide minimal reproduction steps

## Questions?

Open an issue for questions or discussion.
