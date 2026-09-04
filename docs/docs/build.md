# Build and Run

## Tổng quan nhanh

Dix cung cấp các lệnh CLI tích hợp trực tiếp vào luồng phát triển thông thường của Go:

`dix run`

- Dùng khi bạn đang phát triển local.
- Hỗ trợ quét dự án đơn lẻ hoặc toàn bộ Go workspace (`--workspace`), tự động dùng cache để tăng tốc hoặc ép quét lại với `--no-cache`.
- Sau khi generate mã wiring sẽ chạy `go run .`.

`dix build`

- Dùng khi bạn muốn build binary thực thi.
- Hỗ trợ Go workspace (`--workspace`) và cache (`--no-cache`).
- Sau khi generate mã wiring sẽ chạy `go build <target>`.

`dix cache clean`

- Quản lý và dọn dẹp các tệp scan cache được lưu trong `.dix/cache`.

## `dix run [flags] [directory]`

### Cú pháp

```bash
dix run [flags] [directory] [args...]
```

### Tham số & Flags

- `directory` (tùy chọn): Thư mục source để Dix scan. Nếu bỏ qua, mặc định là `.`.
- `args...` (tùy chọn): Các tham số bổ sung sẽ được chuyển tiếp vào lệnh `go run .`.

**Flags:**

- `--workspace`: Quét tất cả các module được khai báo trong tệp `go.work` gần nhất.
- `--no-cache`: Bỏ qua việc đọc và không ghi scan cache (bắt buộc quét lại toàn bộ mã nguồn).

### Ví dụ

```bash
# Quét thư mục hiện tại và chạy
dix run .
```

```bash
# Quét thư mục cụ thể
dix run ./internal/app
```

```bash
# Chạy với Go workspace (monorepo)
dix run --workspace .
```

```bash
# Quét và chạy không dùng cache
dix run --no-cache .
```

```bash
# Truyền thêm tham số cho ứng dụng Go
dix run . --port=8080
```

### Quy trình thực thi

Khi chạy `dix run`, Dix sẽ:

1. Kiểm tra cache scan trong `.dix/cache` (nếu không bật `--no-cache`). Nếu khớp fingerprint, nạp ngay metadata đã cache.
2. Quét source code (hoặc toàn bộ modules trong `go.work` nếu dùng `--workspace`) và parse các function có annotation.
3. Dựng dependency graph từ root.
4. Sinh file `./generated/dix/root.go` (hoặc theo cấu hình `dix.config.json`).
5. Lưu kết quả scan vào cache nếu hợp lệ.
6. Chạy ứng dụng bằng `go run . [args...]`.

## `dix build [flags] [target] [directory]`

### Cú pháp

```bash
dix build [flags] [target] [directory]
```

### Tham số & Flags

- `target` (tùy chọn): Entrypoint file đưa vào `go build`. Mặc định: `main.go`.
- `directory` (tùy chọn): Thư mục source để Dix scan. Mặc định: `.`.

**Flags:**

- `--workspace`: Quét tất cả các module được khai báo trong tệp `go.work` gần nhất.
- `--no-cache`: Bỏ qua việc đọc và không ghi scan cache (bắt buộc quét lại toàn bộ mã nguồn).

### Ví dụ

```bash
# Build với giá trị mặc định (target: main.go, directory: .)
dix build
```

```bash
# Chỉ định rõ target file và thư mục quét
dix build main.go .
```

```bash
# Chỉ định entry file trong cmd và quét thư mục internal
dix build cmd/api/main.go ./internal
```

```bash
# Build ứng dụng trong môi trường Go Workspace
dix build --workspace main.go ./apps/api
```

```bash
# Build không sử dụng cache
dix build --no-cache main.go .
```

### Quy trình thực thi

Khi chạy `dix build`, Dix sẽ:

1. Kiểm tra cache scan trong `.dix/cache` (nếu không bật `--no-cache`). Nếu khớp fingerprint, nạp ngay metadata đã cache.
2. Scan source code theo `directory` (hoặc toàn bộ modules trong `go.work` nếu dùng `--workspace`) và parse các function có annotation.
3. Dựng dependency graph từ root.
4. Sinh file `./generated/dix/root.go` (hoặc theo cấu hình `dix.config.json`).
5. Lưu kết quả scan vào cache nếu hợp lệ.
6. Build ứng dụng bằng `go build <target>`.

Khi build thành công, CLI sẽ in `Build successfully`.

---

## Làm việc với Go Workspace (`--workspace`)

### Tổng quan Go Workspace

Trong các dự án multi-module monorepo sử dụng Go Workspaces (`go.work`), các dependency và provider có thể nằm rải rác ở nhiều Go modules khác nhau (ví dụ: module API chính, module cơ sở dữ liệu dùng chung, module xác thực,...).

Khi truyền cờ `--workspace` cho `dix run`, `dix build` (hoặc `dix wire`), Dix sẽ:

1. **Tìm kiếm file `go.work`**: Bắt đầu từ thư mục scan được chỉ định, Dix tự động duyệt ngược lên các thư mục cha cho đến khi tìm thấy file `go.work`.
2. **Đọc danh sách modules**: Phân tích cú pháp `go.work` và trích xuất tất cả các module được khai báo trong khối `use (...)`.
3. **Kiểm tra tính hợp lệ**: Xác nhận mỗi thư mục module đều tồn tại và chứa file `go.mod`.
4. **Quét đa module với build tag `dix`**: Quét toàn bộ package trong các module với build tag `dix` (nhằm bỏ qua mã nguồn được đánh dấu `//go:build !dix`).
5. **Xây dựng Graph đồng nhất**: Tổng hợp tất cả các provider tìm thấy trong toàn bộ workspace và dựng **một dependency graph duy nhất**.

### Ví dụ cấu trúc dự án Go Workspace

Giả sử dự án monorepo của bạn có cấu trúc như sau:

```text
my-workspace/
├── go.work
├── apps/
│   └── api/
│       ├── go.mod
│       ├── main.go
│       └── server/
│           └── server.go       # Chứa @Root NewServer(...)
└── packages/
    ├── database/
    │   ├── go.mod
    │   └── db.go               # Chứa @Injectable NewDatabase(...)
    └── auth/
        ├── go.mod
        └── auth.go             # Chứa @Injectable NewAuthService(...)
```

Nội dung tệp `go.work`:

```go
go 1.22

use (
	./apps/api
	./packages/database
	./packages/auth
)
```

Lệnh thực thi:

```bash
# Chạy ứng dụng api trong workspace
dix run --workspace ./apps/api

# Build binary cho app api
dix build --workspace main.go ./apps/api
```

### Các quy tắc bắt buộc trong Workspace

> [!IMPORTANT]
> **Quy tắc khi sử dụng Go Workspace:**
> 1. **Duy nhất một `@Root`**: Toàn bộ workspace chỉ được phép có duy nhất **một** provider được đánh dấu `@Root`.
> 2. **Không trùng lặp Provider**: Mỗi kiểu dữ liệu trả về (return type) chỉ được cung cấp bởi **một** hàm `@Injectable` duy nhất trên toàn bộ các module được quét. Nếu có từ hai provider cùng trả về một kiểu dữ liệu, Dix sẽ báo lỗi trùng lặp.
> 3. **Module phải hợp lệ**: File `go.work` phải khai báo ít nhất một module và mỗi module phải có tệp `go.mod` hợp lệ.

---

## Quản lý Cache và Tối ưu tốc độ Build

### Cơ chế hoạt động của Cache

Phân tích cú pháp AST (Abstract Syntax Tree) và tải type-checker của Go cho các dự án lớn có thể tốn thời gian. Để đảm bảo tốc độ phản hồi nhanh nhất, Dix tích hợp sẵn hệ thống **Scan Cache** tự động.

- **Vị trí lưu trữ**: Thư mục `.dix/cache/` (nằm tại thư mục gốc của project đang scan).
- **Khóa Cache (Cache Key / Fingerprint)**: Dix tính toán mã băm SHA-256 dựa trên:
  - Phiên bản schema cache nội bộ (`dix-cache-schema:1`).
  - Phiên bản Go runtime (`runtime.Version()`).
  - Trạng thái cờ `--workspace` (`true` hoặc `false`).
  - Đường dẫn và toàn bộ nội dung của tất cả file `.go`, `go.mod`, `go.sum` trong các thư mục được quét (bỏ qua các thư mục `.git`, `.dix`, `vendor`).
  - Đường dẫn và nội dung của file `go.work` (nếu đang bật chế độ workspace).

### Cache Hit vs Cache Miss

- **Cache Hit**: Nếu không có bất kỳ thay đổi nào trong mã nguồn và file cấu hình kể từ lần chạy trước, Dix sẽ in thông báo:
  ```text
  [Cache] Hit: .dix/cache/a1b2c3d4...json
  ```
  Dix lập tức nạp metadata từ cache và bỏ qua hoàn toàn bước scan cú pháp AST, giúp thời gian khởi động / build gần như tức thì.
- **Cache Invalidation (Tự động hết hạn)**: Cache sẽ tự động bị hủy và Dix sẽ quét lại từ đầu khi:
  - Bất kỳ file Go nào trong dự án bị thay đổi nội dung, thêm mới hoặc xóa.
  - File `go.mod` hoặc `go.sum` thay đổi.
  - File `go.work` thay đổi.
  - Phiên bản Go của hệ thống thay đổi.
  - Chuyển đổi giữa chế độ thông thường và chế độ `--workspace`.

### Bỏ qua Cache (`--no-cache`)

Nếu bạn muốn ép buộc Dix quét lại toàn bộ mã nguồn từ đầu mà không đọc từ cache hiện có và không ghi đè cache mới (thích hợp cho môi trường CI/CD hoặc khi cần debug lại từ trạng thái sạch):

```bash
# Chạy không dùng cache
dix run --no-cache .

# Build không dùng cache
dix build --no-cache main.go .
```

### Lệnh dọn dẹp Cache (`dix cache clean`)

Dix cung cấp lệnh CLI chuyên dụng để xóa sạch thư mục cache `.dix/cache`:

#### Cú pháp

```bash
dix cache clean [directory]
```

- `directory` (tùy chọn): Thư mục chứa thư mục cache `.dix/cache`. Mặc định là thư mục hiện tại (`.`).

#### Ví dụ

```bash
# Xóa cache tại thư mục hiện tại
dix cache clean .

# Xóa cache tại một thư mục dự án cụ thể
dix cache clean ./apps/api
```

Khi xóa thành công, CLI sẽ hiển thị:

```text
[Cache] Removed .dix/cache
```

---

## File được tạo

Cả hai lệnh `dix run` và `dix build` đều quản lý và sinh ra các tệp sau:

- **`./generated/dix/root.go`**: Mã wiring do Dix generate (hoặc đường dẫn được tùy biến trong `dix.config.json` thông qua trường `output`).
- **`scan_<timestamp>.dix`**: Snapshot metadata scan để phục vụ việc theo dõi kết quả phân tích.
- **`.dix/cache/`**: Thư mục lưu các file cache định dạng JSON. Bạn nên thêm `.dix/` vào file `.gitignore` của dự án để tránh commit cache lên repository.
