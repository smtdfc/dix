# Bắt đầu

## Dix là gì?

**Dix** là giải pháp khởi tạo phụ thuộc (Dependency Injection) cho ngôn ngữ Go theo triết lý **"Zero-magic"**. Khác với các thư viện dựa trên Reflection tại Runtime, Dix hoạt động hoàn toàn tại thời điểm biên dịch (Compile-time).

### Giá trị cốt lõi:

- **Compile-time Safety:** Phát hiện lỗi thiếu dependency hoặc vòng lặp phụ thuộc ngay khi sinh mã, loại bỏ rủi ro crash ứng dụng khi khởi động.
- **Hiệu năng tối ưu:** Không sử dụng Reflection, giúp tốc độ thực thi tương đương với việc viết mã khởi tạo bằng tay.
- **Minh bạch & Kiểm soát:** Mã nguồn được sinh ra là Go thuần (Pure Go), cho phép lập trình viên dễ dàng đọc, hiểu và gỡ lỗi (debug).

## Thiết lập môi trường

### Yêu cầu hệ thống

Trước khi triển khai, hãy đảm bảo dự án của bạn đáp ứng các tiêu chuẩn sau:

- Đã cài đặt **Go SDK** (Khuyến nghị phiên bản mới nhất).
- Dự án đã khởi tạo **Go Modules** (`go.mod`).
- Ứng dụng hiện tại có thể thực thi lệnh `go build` mà không có lỗi cú pháp.

### Cài đặt Dix CLI

Sử dụng Go để cài đặt phiên bản mới nhất của công cụ dòng lệnh:

```bash
go install github.com/smtdfc/dix@latest

```

_Lưu ý: Đảm bảo thư mục `$(go env GOPATH)/bin` đã nằm trong biến môi trường `PATH` để có thể gọi lệnh `dix` từ bất kỳ đâu._

## Cách Dix Hoạt động

1. **Scan:** Quét toàn bộ mã nguồn trong thư mục được chỉ định để tìm các hàm khởi tạo có annotation `@Injectable`.
2. **Analyze:** Xác định điểm bắt đầu thông qua annotation `@Root`, từ đó phân tích và dựng đồ thị phụ thuộc (Dependency Graph).
3. **Generate:** Tự động sinh tệp tin `./dix/generated/root.go` chứa hàm `generated.Root()`, đóng vai trò là "nút thắt" kết nối toàn bộ hệ thống.

## Hướng dẫn triển khai thực tế

Giả sử cấu trúc dự án của bạn như sau:

```text
your-app/
├── go.mod
├── main.go
└── internal/
    └── app/
        └── provider.go

```

### Khai báo Provider bằng Annotation

Tại tệp `internal/app/provider.go`, định nghĩa các Constructor rõ ràng. Dix sẽ dựa vào các chú thích sau để thực hiện việc nối mã (wiring):

```go
package app

import "fmt"

type Repo struct{}
type Service struct {
    repo *Repo
}

// @Injectable
// Đăng ký hàm này như một nhà cung cấp dependency
func NewRepo() *Repo {
    return &Repo{}
}

// @Injectable
func NewService(repo *Repo) *Service {
    return &Service{repo: repo}
}

// @Injectable
// @Root
// Đánh dấu đây là Composition Root của ứng dụng
func NewMessage(service *Service) string {
    return "Dix đã kết nối hệ thống thành công!"
}

```

> **Quy tắc quan trọng:**
>
> - Mỗi Provider chỉ nên trả về **duy nhất một giá trị**.
> - Kiểu tham số đầu vào của một Provider phải khớp hoàn toàn với kiểu trả về của Provider tương ứng trong đồ thị.

### Tích hợp vào hàm Main

Sử dụng mã nguồn do Dix sinh ra để khởi tạo ứng dụng trong `main.go`:

```go
package main

import (
    "fmt"
    "github.com/your-org/your-app/internal/app"
    "github.com/your-org/your-app/dix/generated"
)

func main() {
    // Gọi hàm Root được sinh tự động để khởi tạo toàn bộ dependency graph
    msg := generated.Root()
    fmt.Println(msg)
}

```

### Thực thi và Build Binary

Sử dụng lệnh sau để Dix quét mã, sinh code và chạy ứng dụng:

```bash
dix run .

```

Để đóng gói ứng dụng (Build), sử dụng lệnh chuyên dụng:

```bash
dix build main.go .

```

_Trong đó: `main.go` là file entrypoint và `.` là thư mục nguồn để Dix quét._

## Bước tiếp theo

- Tìm hiểu sâu hơn về [Cấu hình nâng cao](https://www.google.com/search?q=/docs/configuration).
- Xem các ví dụ thực tế tại [Dix Github Repository](https://github.com/smtdfc/dix).
