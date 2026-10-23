# Annotations của Dix

Tất cả annotation viết trên một dòng riêng trong doc comment của hàm.

| Annotation | Ý nghĩa |
|------------|---------|
| `@Injectable` | Hàm là một provider hợp lệ trong dependency graph |
| `@Root` | Composition root, điểm bắt đầu duyệt dependency (phải đi cùng `@Injectable`) |
| `@Singleton` | Kết quả provider là singleton, gói trong `di.Singleton[T]` |
| `@Disable` | Tắt provider, không sinh mã và không dùng làm dependency/root |

## Ví dụ đầy đủ

```go
package app

type Database struct{}

type Server struct {
	db *Database
}

// @Injectable
// @Singleton
func NewDatabase() *Database {
	return &Database{}
}

// @Injectable
func NewServer(db *Database) *Server {
	return &Server{db: db}
}

// @Injectable
// @Root
func NewApp(server *Server) string {
	_ = server
	return "app started"
}
```

### `@Singleton`

Khi đánh dấu `@Singleton`, Dix sinh mã gọi provider rồi gói kết quả vào `di.NewSingleton(...)`. Dependencies khác chỉ cần khai báo kiểu `*Database` như thường, Dix sẽ tự giải quyết lấy instance từ singleton.

### `@Disable`

Dùng khi bạn muốn giữ code provider nhưng tạm thời không muốn Dix quét vào graph:

```go
// @Injectable
// @Disable
func OldProvider() *OldService {
	return &OldService{}
}
```

## Quy tắc

- `@Root` phải luôn đi cùng `@Injectable`.
- Một graph chỉ có duy nhất một `@Root`.
- Một kiểu chỉ được có đúng một provider trong toàn bộ module được quét.
- Kiểu dependency phải khớp chính xác: `T` khác `*T`.
- Provider trả về đúng một giá trị, hoặc `(value, error)`.
