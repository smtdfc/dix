# Dix

Zero-magic, compile-time Dependency Injection for Go.

Dix scans your Go source code, builds a dependency graph from annotations in doc comments, and generates clean, explicit wiring code into `generated/dix/root.go`.

Instead of runtime reflection, Dix produces plain Go code that is easy to read, debug, and review.

---

## Key Features

- **Compile-Time Safety**: Catch missing dependencies and circular dependency cycles during code generation—before your application runs.
- **Rich Error Diagnostics**: Beautiful, compiler-grade terminal diagnostics with exact file paths, line & column numbers, source code snippets with caret pointers, and actionable hints.
- **Zero Reflection**: No runtime reflection overhead; wiring performance matches hand-written Go.
- **Explicit Output**: Generates readable, standard Go code that is straightforward to debug, inspect, and trace.
- **Seamless Workflow**: A lightweight CLI that fits directly into your daily `go run` and `go build` routines.
- **Package Ignoring**: Exclude packages you don't need to scan (e.g., mocks, third-party code) with glob and exact patterns to accelerate scans on large codebases.
- **Smart Scan Cache**: Persists parsed metadata in `.dix/cache` using content fingerprints, only rescanning when files actually change.

---

## Installation

```bash
go install github.com/smtdfc/dix@latest
```

Verify your installation:

```bash
dix --help
```

---

## How It Works

1. **Scans** your Go source code for provider functions marked with `@Injectable`.
2. **Identifies** the composition root marked with `@Root`.
3. **Builds** the dependency graph and validates it against missing dependencies and cycles.
4. **Generates** explicit wiring code into `./generated/dix/root.go`.
5. **Executes** `go run .` or `go build <target>` according to your command.

### The `!dix` Build Tag

The generated code includes a build tag header:

```go
//go:build !dix
// +build !dix
```

Dix scans your source code with the `-tags=dix` build flag. When building your application normally without this tag, the generated file is **skipped** (due to `!dix`), preventing symbol collisions while Dix runs during your build pipeline.

---

## Quick Start

### 1. Define Providers

Add doc annotations to your constructor functions:

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

### 2. Use Generated Root in `main.go`

```go
package main

import (
	"fmt"

	"github.com/your-org/your-app/generated/dix"
)

func main() {
	msg, err := dix.Root()
	if err != nil {
		panic(err)
	}
	fmt.Println(msg)
}
```

### 3. Generate Wiring and Run

```bash
dix run .
```

---

## CLI Commands

### `dix run [directory]`

Scans source code, generates wiring, and runs `go run .`.

- `directory` is optional (defaults to `.`).

```bash
dix run .
dix run ./internal/app
dix run --workspace .
dix run . --ignore-package internal/mock
```

### `dix build [target] [directory]`

Generates wiring code and compiles the Go binary.

- `target` is optional (defaults to `main.go`).
- `directory` is optional (defaults to `.`).

```bash
dix build
dix build main.go .
dix build cmd/api/main.go ./internal
dix build --workspace main.go ./apps/api
dix build . --ignore-package github.com/foo/bar
```

### `dix wire [directory]`

Generates wiring code only, without building or running.

- `directory` is optional (defaults to `.`).

```bash
dix wire
dix wire ./internal/app
dix wire --workspace ./apps/api
```

### `dix cache clean [directory]`

Removes all Dix scan cache files (the `.dix/cache` directory).

- `directory` is optional (defaults to `.`).

```bash
dix cache clean .
```

### Global Flags

| Flag | Description | Supported Commands |
|------|-------------|---------------------|
| `--workspace` | Scan every module declared in the nearest `go.work` file | `run`, `build`, `wire` |
| `--no-cache` | Ignore cached metadata and perform a full rescan | `run`, `build`, `wire` |
| `--ignore-package` | Exclude a package ID from scanning (repeatable) | `run`, `build`, `wire` |

---

## Configuration (`dix.config.json`)

Create a `dix.config.json` file in your project root:

```json
{
  "ignore_packages": [
    "github.com/foo/bar/...",
    "internal/mock",
    "github.com/baz/qux/*"
  ],
  "output": "./generated/dix/root.go"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `ignore_packages` | `[]string` | List of package patterns to skip when scanning |
| `output` | `string` | Target path for the generated file (defaults to `./generated/dix/root.go`) |

> [!NOTE]
> Values configured in `dix.config.json` are automatically merged with any `--ignore-package` flags specified via the CLI.

---

## Annotations

Annotations must be placed on their own line inside the provider function's doc comment.

### `@Injectable`

Marks a function as a valid provider within the dependency graph.

### `@Root`

Marks the composition root (the entry point where Dix starts traversing dependencies). Exactly **one** `@Root` is permitted per application, and it must accompany `@Injectable`.

### `@Singleton`

Marks the provider's returned instance as a singleton wrapped inside `di.Singleton[T]`. Dix generates code that calls the provider once and encapsulates it with `di.NewSingleton(...)`.

```go
// @Injectable
// @Singleton
func NewDatabase() *Database {
	return &Database{}
}
```

Inject singletons by value using `di.Singleton[T]`:

```go
// @Injectable
func NewUserRepository(db di.Singleton[*Database]) *UserRepository {
	return &UserRepository{db: db.Get()}
}
```

### `@Disable`

Deactivates a provider. The function will **not** be wired and cannot be used as a dependency or root.

```go
// @Injectable
// @Disable
func OldProvider() *OldService {
	return &OldService{}
}
```

### Important Rules

- `@Root` must always be paired with `@Injectable`.
- Exactly one `@Root` provider is allowed across the entire scanned graph.
- Each type must have only **one** provider (duplicate providers for the same type trigger a validation error).
- Dependency types must match exactly (`T` is distinct from `*T`).
- Provider functions must return either `T` or `(T, error)`. If an error-returning provider fails at startup, the generated code returns the error.

---

## Error Diagnostics

When Dix detects an issue during graph analysis or code generation, it outputs a clean, compiler-grade diagnostic complete with the exact file, line, column, and a source code preview:

```text
[Dix Error] Dependency Resolution Error: provider not found for dependency

  --> internal/app/service.go:4:17
     │
   3 │ // @Injectable
   4 │ func NewMessage(repo *Repo) string {
     │                 ^^^^
   5 │ 	return "hello"
     │

  Provider:   NewMessage
  Dependency: repo (type of *Repo in package app)

  💡 Hint: Ensure a provider function returning this type is defined, annotated with @Injectable, and its package is not ignored in dix.config.json.
```

---

## Ignoring Packages

For large projects, you may want to prevent Dix from scanning mock directories, vendor trees, or legacy code.

### Configuration

In `dix.config.json`:

```json
{
  "ignore_packages": ["internal/mock", "github.com/foo/bar/..."]
}
```

Or via CLI flags (repeatable, merged with config):

```bash
dix run . --ignore-package internal/mock --ignore-package github.com/foo/bar
```

### Supported Patterns

| Pattern | Matches |
|---------|---------|
| `github.com/foo/bar` | Exactly the package `github.com/foo/bar` (does not match sub-packages) |
| `github.com/foo/bar/...` | Package `bar` and all descendant sub-packages |
| `github.com/foo/bar/*` | Package `bar` and its immediate children `bar/*` |
| `github.com/foo/*bar*` | Glob pattern (`*` matches any sequence of non-slash characters) |
| `internal/mock` | Exact match for `internal/mock` |

When one or more packages are ignored, Dix outputs a clean summary:

```text
[Ignore] 3 package(s) ignored, 42 scanned
```

---

## Scan Cache

Dix stores validated metadata in `.dix/cache/<fingerprint>.json`.

The cache fingerprint is computed from:
- Go version (`runtime.Version()`)
- Scan mode (`--workspace` status)
- Contents of `.go`, `go.mod`, and `go.sum` files within scanned directories
- The `go.work` file if using `--workspace`

The cache is automatically invalidated when any of these inputs change.

- **Force a rescan**: Pass the `--no-cache` flag.
- **Clear cache**: Run `dix cache clean .`.

---

## Generated Artifacts

| File | Description |
|------|-------------|
| `generated/dix/root.go` | The generated wiring code containing the `Root()` entry function |
| `.dix/scan_<timestamp>.dix` | Intermediate metadata snapshot from the scan |
| `.dix/cache/<hash>.json` | Cached scan metadata |

---

## Common Errors & Troubleshooting

| Error | Cause & Solution |
|-------|------------------|
| `multiple @Root providers` | Two or more functions are marked with `@Root`. Keep `@Root` on only one entry point. |
| `duplicate provider for ...` | Two providers return the same type. Remove one, rename the type, or mark the unused one with `@Disable`. |
| `provider not found for dependency` | A dependency is missing a provider. Ensure a provider function returns the required type and has the `@Injectable` annotation. |
| `circular dependency detected` | A dependency loop was found. Restructure your dependencies or use `di.Singleton[T]` to resolve the cycle. |
| `singleton dependency must be di.Singleton[T]` | Singletons must be passed by value (`di.Singleton[T]`), not as a pointer (`*di.Singleton[T]`). |
| `failed to load Go packages` | Go packages failed to load (usually due to a syntax error or missing module). Check your code with `go build`. |

---

## Contributing

```bash
git clone https://github.com/smtdfc/dix.git
cd dix
go test ./...
```

Contributions, bug reports, and pull requests are welcome!

---

## Documentation

| Guide | File |
|-------|------|
| Getting Started | [`docs/getting-started.md`](docs/getting-started.md) |
| Installation | [`docs/installation.md`](docs/installation.md) |
| Annotations | [`docs/annotations.md`](docs/annotations.md) |
| Build & Run | [`docs/build.md`](docs/build.md) |
| Scan Cache | [`docs/cache.md`](docs/cache.md) |
| Package Ignoring | [`docs/ignore.md`](docs/ignore.md) |
| Common Errors | [`docs/common-errors.md`](docs/common-errors.md) |

---

## License

MIT License. See [LICENSE](LICENSE) for details.
