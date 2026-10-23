# Hướng dẫn bắt đầu với Dix

Dix là công cụ sinh mã Dependency Injection compile-time cho Go, không dùng reflection tại runtime.

## 1. Cài đặt

```bash
go install github.com/smtdfc/dix@latest
dix --help
```

## 2. Định nghĩa provider

Tạo các hàm provider và đánh dấu bằng doc comment `@Injectable`:

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

## 3. Sinh mã và chạy

```bash
dix run .
```

Dix sẽ:
1. Quét source tìm `@Injectable` và `@Root`.
2. Sinh file `dix/generated/root.go` chứa hàm `Root()`.
3. Chạy `go run .`.

Trong `main.go`:

```go
package main

import (
	"fmt"

	"github.com/your-org/your-app/dix/generated"
)

func main() {
	msg, err := generated.Root()
	if err != nil {
		panic(err)
	}
	fmt.Println(msg)
}
```

## 4. Các lệnh thường dùng

| Lệnh | Ý nghĩa |
|------|---------|
| `dix run .` | Quét, sinh mã, rồi chạy `go run .` |
| `dix build` | Quét, sinh mã, rồi `go build main.go` |
| `dix wire` | Chỉ sinh mã wiring, không build/chạy |
| `dix cache clean .` | Xóa scan cache ở `.dix/cache` |

## 5. Tiếp theo

Xem chi tiết:
- [Annotations](./annotations.md) - `@Injectable`, `@Root`, `@Singleton`, `@Disable`
- [Cấu hình](./ignore.md) - `dix.config.json` và `--ignore-package`
- [Cache](./cache.md) - `.dix/cache` và `--no-cache`
- [Lỗi thường gặp](./common-errors.md)
