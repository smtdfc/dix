# Build và Run

## `dix run [directory]`

Quét source, sinh mã wiring, rồi thực thi `go run .`.

```bash
dix run .
dix run ./internal/app
dix run --workspace .
dix run . --ignore-package internal/mock
```

## `dix build [target] [directory]`

Kết hợp sinh mã và compile binary.

```bash
dix build                 # go build main.go
dix build main.go .
dix build cmd/api/main.go ./internal
dix build --workspace main.go ./apps/api
dix build . --ignore-package github.com/foo/bar
```

- `target` mặc định là `main.go`.
- `directory` mặc định là `.`.

## `dix wire [directory]`

Chỉ sinh mã wiring vào `dix/generated/root.go`, không build và không chạy.

```bash
dix wire
dix wire ./internal/app
dix wire --workspace ./apps/api
```

## Build tag `!dix`

Mã sinh ra có header:

```go
//go:build !dix
// +build !dix
```

Dix quét source với `-tags=dix`. Khi bạn build dự án bình thường (không thêm tag `dix`), file generated sẽ được bỏ qua, tránh trùng ký hiệu với provider thường.

## Flags chung

| Flag | Ý nghĩa |
|------|---------|
| `--workspace` | Quét mọi module trong `go.work` gần nhất |
| `--no-cache` | Bỏ qua cache, quét lại từ đầu |
| `--ignore-package` | Bỏ qua một package (có thể lặp nhiều lần) |
