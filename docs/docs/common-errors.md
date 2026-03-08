# Lỗi thường gặp

Trang này tổng hợp các lỗi phổ biến khi dùng Dix và cách xử lý nhanh.

## Nhóm Parser

`parser/validation: provider function must return exactly one value`

- Nguyên nhân: hàm `@Injectable` không return hoặc return nhiều hơn một giá trị.
- Cách sửa: đảm bảo provider chỉ return đúng một giá trị.

`parser/validation: singleton dependency must be di.Singleton[T], not *di.Singleton[T]`

- Nguyên nhân: dùng sai dạng `*di.Singleton[T]`.
- Cách sửa: đổi sang `di.Singleton[T]`.

`parser/package_load: failed to load Go packages`

- Nguyên nhân: source không compile được theo context parser hoặc package lỗi.
- Cách sửa: kiểm tra lỗi Go compiler/type checker ở package được báo.

## Nhóm Generate / Graph

`generator/validation: cannot find @Root provider`

- Nguyên nhân: chưa có provider được đánh dấu `@Root` hợp lệ.
- Cách sửa: thêm `@Injectable` + `@Root` cho đúng một hàm root.

`generator/graph_build: provider must declare a return value`

- Nguyên nhân: provider bị thiếu return type hoặc metadata không hợp lệ.
- Cách sửa: sửa chữ ký provider để có return hợp lệ.

`generator/dependency_resolution: singleton dependency provider not found`

- Nguyên nhân: dependency singleton không có provider trả về kiểu tương ứng.
- Cách sửa: thêm provider đúng type.

`generator/dependency_resolution: dependency is not available in generated container scope`

- Nguyên nhân: dependency thường chưa được tạo trong scope generate.
- Cách sửa: kiểm tra graph và type matching giữa provider return/param.

`generator/dependency_resolution: circular dependency detected [provider=<ProviderName>]`

- Nguyên nhân: graph dependency bị vòng (ví dụ `A -> B -> C -> A`) nên không thể sắp xếp thứ tự khởi tạo.
- Cách sửa:
  1.  Tách một dependency trong vòng thành abstraction khác.
  2.  Dời phần phụ thuộc chéo sang method/runtime wiring thay vì constructor injection trực tiếp.
  3.  Kiểm tra lại type return/param để tránh vô tình tạo vòng qua interface tương đương.

`generator/dependency_resolution: disabled provider cannot be used as dependency or root [provider=<ProviderName>]`

- Nguyên nhân: một provider được đánh dấu `@Disable` nhưng vẫn nằm trong dependency chain từ `@Root`.
- Cách sửa:
  1.  Bỏ `@Disable` nếu provider đó vẫn cần dùng.
  2.  Hoặc thay dependency bằng provider khác còn hoạt động.
  3.  Nếu provider đã deprecated, tách root/dependency graph để không còn tham chiếu vào provider này.

## Nhóm Build và Run

`Error: exit status 1`

- Nguyên nhân: bước `go run` hoặc `go build` của project thất bại sau khi đã generate.
- Cách sửa: đọc log compiler/runtime ngay phía trên để xử lý lỗi ứng dụng.

## Gợi ý debug nhanh

1. Chạy `dix run .` để xem lỗi parser/generate trước.
2. Nếu parser báo type mismatch, so sánh chính xác kiểu `T` và `*T`.
3. Nếu graph báo thiếu provider, kiểm tra provider đó đã có `@Injectable` chưa.
