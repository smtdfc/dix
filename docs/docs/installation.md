# Cài đặt Dix

Trang này hướng dẫn cài đặt `dix` CLI cho môi trường phát triển Go.

## Yêu cầu

Trước khi cài đặt, đảm bảo bạn đã có:

- Go đã được cài (`go version` chạy được)
- Kết nối mạng để tải module từ GitHub
- Quyền ghi vào thư mục cài binary của Go (`GOBIN` hoặc `$(go env GOPATH)/bin`)

## Cài đặt nhanh (khuyến nghị)

Đây là cách đơn giản nhất để cài phiên bản mới nhất:

```bash
go install github.com/smtdfc/dix@latest
```

Sau khi cài xong, kiểm tra:

```bash
dix --help
```

Nếu thấy danh sách lệnh CLI là cài đặt thành công.

## Cài đặt phiên bản cụ thể

Nếu bạn muốn pin version cho môi trường team/CI:

```bash
go install github.com/smtdfc/dix@v0.1.0
```

Thay `v0.1.0` bằng phiên bản bạn muốn sử dụng.

## Cài từ source (dành cho contributor)

Khi cần chỉnh sửa Dix hoặc test nhánh riêng:

```bash
git clone https://github.com/smtdfc/dix.git
cd dix
go install .
```

Lệnh `go install .` sẽ build từ mã local và cài binary `dix` vào đường dẫn cài đặt của Go.

## Kiểm tra đường dẫn binary

`go install` đặt binary tại:

- `$(go env GOBIN)` nếu `GOBIN` đã được cấu hình
- Nếu chưa cấu hình `GOBIN`: `$(go env GOPATH)/bin`

Bạn có thể kiểm tra nhanh:

```bash
go env GOBIN
go env GOPATH
```

Nếu chạy `dix` báo "command not found", hãy thêm đường dẫn binary vào `PATH`.

## Cấu hình PATH (Linux/macOS)

Thêm vào shell profile (`~/.bashrc`, `~/.zshrc`, ...):

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Sau đó nạp lại cấu hình shell:

```bash
source ~/.bashrc
```

Nếu bạn dùng `zsh`, thay bằng `source ~/.zshrc`.

## Cập nhật Dix

Để cập nhật lên bản mới nhất:

```bash
go install github.com/smtdfc/dix@latest
```

## Gỡ cài đặt

Xóa binary `dix` khỏi thư mục cài của Go:

```bash
rm -f "$(go env GOPATH)/bin/dix"
```

Nếu bạn dùng `GOBIN`, xóa file tương ứng trong thư mục đó.

## Lỗi thường gặp khi cài đặt

`dix: command not found`

- Nguyên nhân: thư mục chứa binary chưa có trong `PATH`.
- Cách xử lý: thêm `$(go env GOPATH)/bin` (hoặc `GOBIN`) vào `PATH`.

`permission denied` khi cài

- Nguyên nhân: không có quyền ghi vào thư mục cài binary.
- Cách xử lý: cấu hình `GOBIN` về thư mục user có quyền ghi.

`unknown revision` khi dùng phiên bản cụ thể

- Nguyên nhân: tag version không tồn tại.
- Cách xử lý: kiểm tra lại release/tag hợp lệ trên GitHub.
