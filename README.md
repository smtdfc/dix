# Dix

Zero-magic, compile-time Dependency Injection for Go.

Dix scans your Go source code, builds a dependency graph from annotations in doc comments, and generates wiring code at `dix/generated/root.go`.

Instead of runtime reflection, Dix produces plain Go code that is easy to read, debug, and review.

## Why Dix?

- Compile-time safety: catch missing dependencies and circular dependency loops during generation.
- No reflection overhead: runtime performance stays close to handwritten wiring.
- Explicit output: generated code is transparent and maintainable.
- Simple workflow: a small CLI that fits naturally into `go run` and `go build` flows.

## Installation

```bash
go install github.com/smtdfc/dix@latest
```

Verify installation:

```bash
dix --help
```

## How It Works

1. Scan source code for providers marked with `@Injectable`.
2. Locate the composition root marked with `@Root`.
3. Build a dependency graph.
4. Generate `dix/generated/root.go`.
5. Run `go run .` or `go build <target>` depending on the command you choose.

## Quick Start

### 1. Define providers

```go
package app

type Repo struct{}

type Service struct {
	repo *Repo
}

// @Injectable
func NewRepo() *Repo {
	return &Repo{}
}

// @Injectable
func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

// @Injectable
// @Root
func NewMessage(service *Service) string {
	_ = service
	return "Dix wiring done"
}
```

### 2. Use generated root in `main.go`

```go
package main

import (
	"fmt"

	"github.com/your-org/your-app/dix/generated"
)

func main() {
	msg := generated.Root()
	fmt.Println(msg)
}
```

### 3. Generate and run

```bash
dix run .
```

## CLI Commands

### `dix run [directory]`

- Parses source, generates wiring, then executes `go run .`.
- `directory` is optional and defaults to `.`.

Examples:

```bash
dix run .
dix run ./internal/app
```

### `dix build [target] [directory]`

- Parses source, generates wiring, then executes `go build <target>`.
- `target` is optional and defaults to `main.go`.
- `directory` is optional and defaults to `.`.

Examples:

```bash
dix build
dix build main.go .
dix build cmd/api/main.go ./internal
```

## Annotations

### `@Injectable`

Marks a function as a valid provider in the dependency graph.

### `@Root`

Marks the composition root where Dix starts traversing dependencies.

Important rules:

- `@Root` must be used together with `@Injectable`.
- A graph should have one root.
- A provider should return exactly one value.
- Dependency types must match exactly (`T` is different from `*T`).

## Generated Artifacts

- `dix/generated/root.go`: generated wiring code.
- `scan_<timestamp>.dix`: scan metadata artifact.

## Documentation

- Getting started: `docs/docs/getting-started.md`
- Installation: `docs/docs/installation.md`
- Annotations: `docs/docs/annotations.md`
- Build and run: `docs/docs/build.md`
- Common errors: `docs/docs/common-errors.md`

## Contributing

```bash
git clone https://github.com/smtdfc/dix.git
cd dix
go test ./...
```

Open an issue or pull request if you want to improve features or docs.

## License

MIT. See `LICENSE`.
