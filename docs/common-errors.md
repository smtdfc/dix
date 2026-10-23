# Lỗi thường gặp

| Lỗi | Nguyên nhân | Cách xử lý |
|-----|-------------|------------|
| `multiple @Root providers: ... and ...` | Hai hàm trở lên đánh dấu `@Root` trong cùng graph. | Giữ lại chỉ một `@Root`, đánh dấu `@Disable` cho các hàm còn lại. |
| `duplicate provider for ...` | Hai provider trả về cùng kiểu trong cùng graph. | Xóa hoặc `@Disable` một trong hai provider. |
| `provider not found for dependency` | Không tìm thấy provider cho một kiểu dependency. | Thêm provider `@Injectable` cho kiểu đó, hoặc kiểm tra kiểu `T` vs `*T`. |
| `circular dependency detected` | Có chu trình dependency trong graph. | Thiết kế lại để loại bỏ vòng tròn. |
| `singleton dependency must be di.Singleton[T], not *di.Singleton[T]` | Khai báo dependency singleton sai định dạng. | Dùng `di.Singleton[T]` thay vì `*di.Singleton[T]`. |
| `failed to load Go packages` | Lỗi load package Go (cú pháp, module, ...). | Sửa lỗi compile dự án, sau đó chạy lại Dix. |
| `provider function must return one value or (value, error)` | Provider trả về sai số lượng giá trị. | Điều chỉnh hàm trả về đúng một giá trị hoặc `(T, error)`. |
| `the second provider return value must be error` | Giá trị thứ hai trả về không phải kiểu `error`. | Đổi thành `error`. |

## Xem chi tiết lỗi

Dix in lỗi kèm theo nguyên nhân (nếu có) để bạn dễ dàng định vị:

```
[Error]
generator/dependency_resolution: provider not found for dependency [provider=NewService] [depends_on= *Repo@github.com/...]
```
