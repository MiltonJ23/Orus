# Development Guide

## Prerequisites

- Go >= 1.25
- Platform-specific Gio dependencies (see [README](../../README.md#prerequisites))

## Build Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/orus` |
| `make run` | Build and run the application |
| `make clean` | Remove build artifacts |
| `make test` | Run all tests |
| `make fmt` | Format all Go source files |
| `make test-with-coverage` | Generate `coverage.out` |
| `make consult-coverage` | View coverage report in terminal |
| `make lint` | Run golangci-lint |
| `make vet` | Run `go vet` |

## Project Conventions

### Code Style

- Follow standard Go formatting (`gofmt`).
- All public types, functions, and methods must have godoc comments.
- Error messages use lowercase and wrap with `fmt.Errorf("context: %w", err)`.
- Use `context.Context` as the first parameter for all I/O operations.

### Naming

- **Packages:** lowercase, single word (`domain`, `service`, `port`).
- **Interfaces:** verb or noun describing the capability (`BookRepository`, `ContentReader`).
- **Errors:** `Err` prefix with descriptive name (`ErrBookNotFound`, `ErrInvalidRating`).
- **Factory functions:** `New` prefix (`NewBook`, `NewReminder`).

### Testing

- Unit tests live alongside the code they test (`book_test.go` next to `book.go`).
- Service tests use mock implementations of port interfaces.
- Table-driven tests are preferred for multiple scenarios.
- Run tests with `make test` before submitting a PR.

### Architecture Rules

1. **Domain** depends on nothing (no imports from `port/`, `service/`, or `adapters/`).
2. **Ports** depend only on `domain/`.
3. **Services** depend on `domain/` and `port/`.
4. **Adapters** depend on `domain/`, `port/`, and external libraries.
5. **cmd/** wires everything together (dependency injection).

### Branching Strategy

```
feat/* → dev → stage → main
```

- `feat/*` — feature branches, branched from `dev`
- `dev` — integration branch (unit tests + linting)
- `stage` — beta/RC (full E2E test suite)
- `main` — production releases only

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add EPUB table of contents extraction
fix: resolve nil pointer in tracker service
docs: update architecture decision records
style: adjust sidebar gold accent color
```

## CI/CD Pipeline

### On Push/PR to `dev`

- Runs unit tests on Linux, macOS, and Windows
- Runs `go vet` and `golangci-lint`
- Checks code formatting

### On Tag Push (`v*`)

- Builds release binaries for all platforms
- Creates a GitHub Release with attached binaries
- Generates changelog from commit history
