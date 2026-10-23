# Dix

Zero-magic, compile-time Dependency Injection for Go.

Dix scans your Go source code, builds a dependency graph from annotations in doc comments, and generates wiring code at `dix/generated/root.go`.

Instead of runtime reflection, Dix produces plain Go code that is easy to read, debug, and review.

## Tính năng chính

- **Compile-time safety**: phát hiện thiếu dependency và vòng tròn dependency ngay lúc sinh mã, không chờ chạy chương trình.
- **Zero reflection**: không có chi phí runtime, tốc độ gần với mã viết tay.
- **Explicit output**: mã sinh ra là Go thuần sẽ đọc và debug dễ dàng.
- **Workflow đơn giản**: một CLI nhỏ phù hợp vào luồng `go run` và `go build`.
- **Ignore packages**: bỏ qua các package không cần scan (giúp scan nhanh hơn trên project lớn).
- **Scan cache**: lưu metadata đã scan vào `.dix/cache`, chỉ quét lại khi source thay đổi.

## Cài đặt

```bash
go install github.com/smtdfc/dix@latest
```

Kiểm tra:

```bash
dix --help
```

## Nguyên lý hoạt động

1. Quét source code tìm các hàm đánh dấu `@Injectable`.
2. Xác định composition root đánh dấu `@Root`.
3. Xây dựng dependency graph và kiểm tra vòng tròn.
4. Sinh mã wiring vào `dix/generated/root.go`.
5. Chạy `go run .` hoặc `go build <target>` tùy lệnh bạn chọn.

### Build tag `!dix`

Mã sinh ra có header:

```go
//go:build !dix
// +build !dix
```

Dix quét source với build tag `-tags=dix`. Nếu bạn không thêm tag này khi build dự án, file generated sẽ **bị bỏ qua** (do `!dix`) và không gây xung đột với các provider định nghĩa thường xuyên. Điều này tránh trùng ký hiệu khi Dix chạy trong quá trình build dự án.

## Quick Start

### 1. Định nghĩa provider

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

### 2. Dùng generated root trong `main.go`

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

### 3. Sinh mã và chạy

```bash
dix run .
```

## Lệnh CLI

### `dix run [directory]`

Quét source, sinh mã wiring, rồi thực thi `go run .`.

- `directory` là tùy chọn, mặc định `.`.

Ví dụ:

```bash
dix run .
dix run ./internal/app
dix run --workspace .
dix run . --ignore-package internal/mock
```

### `dix build [target] [directory]`

Kết hợp sinh mã và compile binary.

- `target` là tùy chọn, mặc định `main.go`.
- `directory` là tùy chọn, mặc định `.`.

Ví dụ:

```bash
dix build
dix build main.go .
dix build cmd/api/main.go ./internal
dix build --workspace main.go ./apps/api
dix build . --ignore-package github.com/foo/bar
```

### `dix wire [directory]`

Chỉ sinh mã wiring, không build và không chạy.

- `directory` là tùy chọn, mặc định `.`.

Ví dụ:

```bash
dix wire
dix wire ./internal/app
dix wire --workspace ./apps/api
```

### `dix cache clean [directory]`

Xóa toàn bộ scan cache của Dix (thư mục `.dix/cache`).

- `directory` là tùy chọn, mặc định `.`.

Ví dụ:

```bash
dix cache clean .
```

### Danh sách flag chung

| Flag | Ý nghĩa | Áp dụng cho |
|------|---------|-------------|
| `--workspace` | Quét mọi module trong `go.work` gần nhất | `run`, `build`, `wire` |
| `--no-cache` | Bỏ qua cache, quét lại từ đầu | `run`, `build`, `wire` |
| `--ignore-package` | Bỏ qua một package theo ID (có thể lặp nhiều lần) | `run`, `build`, `wire` |

## Cấu hình `dix.config.json`

Tạo file `dix.config.json` tại thư mục gốc dự án:

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

| Trường | Kiểu | Ý nghĩa |
|--------|------|---------|
| `ignore_packages` | `[]string` | Danh sách package bỏ qua khi scan |
| `output` | `string` | Đường dẫn file generated (mặc định `./generated/dix/root.go`) |

Giá trị từ file config sẽ được **ghép** với giá trị từ flag `--ignore-package`.

## Annotations

Tất cả annotation đều viết trên một dòng riêng trong doc comment của hàm.

### `@Injectable`

Đánh dấu hàm là một provider hợp lệ trong dependency graph.

### `@Root`

Đánh dấu composition root (điểm bắt đầu Dix duyệt dependency). Một graph chỉ có đúng **một** `@Root`, và phải đi kèm với `@Injectable`.

### `@Singleton`

Đánh dấu kết quả của provider là singleton, được bọc trong `di.Singleton[T]`. Khi một dependency cần singleton, Dix sẽ sinh mã gọi provider rồi gói vào `di.NewSingleton(...)`.

Ví dụ:

```go
// @Injectable
// @Singleton
func NewDatabase() *Database {
	return &Database{}
}
```

### `@Disable`

Đánh dấu provider bị tắt. Provider này sẽ **không** được sinh mã và không thể dùng làm dependency hay root.

Ví dụ:

```go
// @Injectable
// @Disable
func OldProvider() *OldService {
	return &OldService{}
}
```

### Quy tắc quan trọng

- `@Root` phải luôn đi cùng `@Injectable`.
- Một graph chỉ có duy nhất một `@Root`.
- Một kiểu chỉ được có đúng một provider trong toàn bộ các module được quét.
- Kiểu dependency phải khớp chính xác (`T` khác `*T`).
- Provider trả về đúng một giá trị, hoặc `(value, error)`. Đối với dạng trả lỗi, sinh mã sẽ panic nếu error không nil.

## Bỏ qua package (`ignore packages`)

Khi dự án lớn, bạn có thể không muốn Dix quét các package như `internal/mock`, thư viện của bên thứ ba, hoặc mã cũ.

### Cách cấu hình

Trong `dix.config.json`:

```json
{
  "ignore_packages": ["internal/mock", "github.com/foo/bar/..."]
}
```

Hoặc qua flag CLI (lặp được, ghép với config):

```bash
dix run . --ignore-package internal/mock --ignore-package github.com/foo/bar
```

### Các mẫu pattern được hỗ trợ

| Mẫu | Khớp với |
|-----|----------|
| `github.com/foo/bar` | Chính xác gói `github.com/foo/bar` (không khớp sub-package) |
| `github.com/foo/bar/...` | Gói `bar` và mọi gói con bên dưới |
| `github.com/foo/bar/*` | Chính xác `bar` và gói con trực tiếp `bar/*` |
| `github.com/foo/*bar*` | Glob (dấu `*` đại diện cho mọi ký tự trừ `/`) |
| `internal/mock` | Chính xác `internal/mock` (không làm fuzzy match) |

> Lưu ý: Pattern thuần túy giờ đây là **exact match** (không dùng `strings.Contains` nữa), nên sẽ không bị false-positive như `"foo"` khớp `"foobar"`.

Khi có ít nhất một package bị bỏ qua, Dix sẽ in tóm tát thay vì spam log từng gói:

```
[Ignore] 3 package(s) ignored, 42 scanned
```

## Scan cache

Dix lưu metadata đã quét thành công vào `.dix/cache/<fingerprint>.json`.

Fingerprint được tính từ:
- Phiên bản Go (`runtime.Version()`)
- Chế độ scan (`--workspace` hay không)
- Nội dung các file `.go`, `go.mod`, `go.sum` trong thư mục quét
- File `go.work` nếu dùng `--workspace`

Cache bị hủy và quét lại khi một trong những yếu tố trên thay đổi.

- Ép quét lại: dùng flag `--no-cache`.
- Xóa toàn bộ cache: `dix cache clean .`.

## Kết quả sinh ra

| File | Ý nghĩa |
|------|---------|
| `dix/generated/root.go` | Mã wiring được sinh ra (chứa hàm `Root()`) |
| `scan_<timestamp>.dix` | Tệp metadata intermediate từ lần quét |

## Lỗi thường gặp

| Lỗi | Nguyên nhân |
|-----|-------------|
| `multiple @Root providers` | Hai hàm trở lên đánh dấu `@Root`. |
| `duplicate provider for ...` | Hai provider trả về cùng kiểu trong cùng graph. |
| `provider not found for dependency` | Không tìm thấy provider cho một kiểu dependency. |
| `circular dependency detected` | Có chu trình dependency trong graph. |
| `singleton dependency must be di.Singleton[T]` | Dependency singleton sai định dạng. |
| `failed to load Go packages` | Lỗi load package Go (thường do lỗi cú pháp hoặc module). |

Tất cả lỗi được in ra với stack nguyên nhân (nếu có) để bạn dễ dàng định vị nguồn gốc.

## Đóng góp

```bash
git clone https://github.com/smtdfc/dix.git
cd dix
go test ./...
```

Mở issue hoặc pull request nếu bạn muốn cải thiện tính năng hoặc tài liệu.

## Tài liệu chi tiết

| Chủ đề | File |
|--------|------|
| Hướng dẫn bắt đầu | `docs/getting-started.md` |
| Cài đặt | `docs/installation.md` |
| Annotations | `docs/annotations.md` |
| Build và Run | `docs/build.md` |
| Scan Cache | `docs/cache.md` |
| Bỏ qua package | `docs/ignore.md` |
| Lỗi thường gặp | `docs/common-errors.md` |

## Giấy phép

MIT. Xem `LICENSE`.
