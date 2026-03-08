# Build and Run

## Tổng quan nhanh

`dix run`

- Dùng khi bạn đang phát triển local.
- Sau khi generate sẽ chạy `go run .`.

`dix build`

- Dùng khi bạn muốn build binary.
- Sau khi generate sẽ chạy `go build <target>`.

## `dix run [directory]`

### Cú pháp

```bash
dix run [directory]
```

### Tham số

- `directory` (tùy chọn): thư mục source để Dix scan.
- Nếu bỏ qua, mặc định là `.`.

### Ví dụ

```bash
dix run .
```

```bash
dix run ./internal/app
```

### Quy trình thực thi

Khi chạy `dix run`, Dix sẽ:

1. Scan source code và parse các function có annotation.
2. Dựng dependency graph từ root.
3. Sinh file `./dix/generated/root.go`.
4. Chạy ứng dụng bằng `go run .`.

## `dix build [target] [directory]`

### Cú pháp

```bash
dix build [target] [directory]
```

### Tham số

- `target` (tùy chọn): entry file đưa vào `go build`.
- `directory` (tùy chọn): thư mục source để Dix scan.

Giá trị mặc định:

- `target`: `main.go`
- `directory`: `.`

### Ví dụ

```bash
dix build
```

```bash
dix build main.go .
```

```bash
dix build cmd/api/main.go ./internal
```

### Quy trình thực thi

Khi chạy `dix build`, Dix sẽ:

1. Scan source code theo `directory`.
2. Dựng dependency graph.
3. Sinh file `./dix/generated/root.go`.
4. Build bằng `go build <target>`.

Khi build thành công, CLI sẽ in `Build successfully`.

## File được tạo

Cả hai lệnh đều sinh:

- `dix/generated/root.go`: mã wiring do Dix generate.

File generated này được thêm build tag ở đầu file:

```go
//go:build !dix
// +build !dix
```

Ý nghĩa:

- Khi Dix scan với `-tags=dix`, file generated sẽ bị bỏ qua để tránh tự quét lại và gây nhiễu lỗi.
- Khi chạy/built ứng dụng bình thường (không bật tag `dix`), file generated vẫn được sử dụng như bình thường.

Ngoài ra Dix cũng lưu metadata scan với tên dạng `scan_<timestamp>.dix` để phục vụ việc theo dõi kết quả phân tích.
