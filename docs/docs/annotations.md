# Hệ thống Annotations trong Dix

## Tổng quan về Annotations

Trong Dix, Annotations không phải là tính năng của ngôn ngữ Go mà là các chỉ thị (directives) được đặt trong phần chú thích (Doc Comments). Parser của Dix sẽ quét AST (Abstract Syntax Tree) để trích xuất các chỉ thị này, từ đó xác định cách thức kết nối các thành phần trong dự án.

Hiện tại, Dix hỗ trợ ba annotation chính:

- `@Injectable`: Đăng ký thành phần vào hệ thống DI.
- `@Root`: Xác định điểm khởi tạo gốc của ứng dụng.
- `@Disable`: Tạm thời loại provider khỏi đồ thị phụ thuộc.

## Chi tiết các Annotations

### 1. @Injectable

`@Injectable` là annotation cơ bản nhất, dùng để đánh dấu một hàm Constructor là một **Provider** hợp lệ bên trong Dependency Graph.

#### Mục đích:

Thông báo cho Dix rằng kết quả trả về của hàm này có thể được dùng làm tham số đầu vào cho các hàm khác trong hệ thống.

#### Ví dụ tiêu chuẩn:

```go
// @Injectable
// NewRepository khởi tạo tầng lưu trữ dữ liệu
func NewRepository() *Repository {
    return &Repository{}
}

```

#### Quy tắc nhận diện của Parser:

- **Tính nhất quán:** Annotation phân biệt chữ hoa - chữ thường (Case-sensitive). Phải viết chính xác là `@Injectable`.
- **Vị trí ưu tiên:** Để đảm bảo Parser nhận diện ổn định, `@Injectable` nên được đặt ở **dòng đầu tiên** của khối chú thích ngay phía trên hàm.
- **Phạm vi:** Chỉ áp dụng cho các hàm (functions), không áp dụng trực tiếp lên struct hay biến.

### 2. @Root

`@Root` định nghĩa **Composition Root** — điểm nút cao nhất nơi toàn bộ đồ thị phụ thuộc được hội tụ.

#### Mục đích:

Xác định hàm mục tiêu mà từ đó Dix sẽ truy ngược (backtrack) để tìm kiếm và khởi tạo toàn bộ các phụ thuộc cần thiết.

#### Ví dụ tiêu chuẩn:

```go
// @Injectable
// @Root
// NewApp hội tụ các service để khởi chạy ứng dụng
func NewApp(svc *Service) *App {
    return &App{Service: svc}
}

```

#### Quy tắc quan trọng:

- **Tính bắt buộc:** Một hàm được đánh dấu `@Root` **phải** đi kèm với `@Injectable`. Dix chỉ bắt đầu quét các hàm có `@Root` sau khi đã xác nhận chúng là một Provider hợp lệ.
- **Duy nhất:** Mỗi đồ thị phụ thuộc chỉ được phép có một điểm `@Root` duy nhất để tránh xung đột logic khởi tạo.
- **Vị trí:** Phải nằm trên một dòng riêng biệt trong khối Doc Comment.

## Tiêu chuẩn Chữ ký Hàm (Function Signature)

Khi sử dụng Annotations, các hàm Constructor phải tuân thủ nghiêm ngặt các ràng buộc về mặt kỹ thuật để Dix có thể sinh mã chính xác:

1. **Giá trị trả về đơn (Single Return Value):** Mỗi Provider chỉ được phép trả về đúng một giá trị. Nếu hàm của bạn trả về lỗi (error), hãy xử lý hoặc bọc (wrap) lỗi đó trước khi đưa vào luồng khởi tạo của Dix.
2. **Khớp kiểu dữ liệu (Type Matching):** Kiểu dữ liệu trả về của Provider A phải khớp hoàn toàn với kiểu dữ liệu tham số đầu vào của Provider B (bao gồm cả việc phân biệt giữa con trỏ `*T` và giá trị `T`).
3. **Tính hiển thị (Visibility):** Các hàm và kiểu dữ liệu nên được Export (viết hoa chữ cái đầu) nếu chúng nằm ở các package khác nhau để đảm bảo mã sinh ra trong thư mục `./dix/generated` có thể truy cập được.

### 3. @Disable

`@Disable` cho phép đánh dấu một provider là "không được phép sử dụng" trong quá trình dựng graph.

#### Mục đích:

Hữu ích khi bạn muốn giữ lại code provider để refactor hoặc thử nghiệm, nhưng không muốn Dix dùng provider đó để wiring.

#### Ví dụ:

```go
// @Injectable
// @Disable
func NewLegacyRepo() *LegacyRepo {
    return &LegacyRepo{}
}
```

#### Hành vi:

- Provider có `@Disable` vẫn được parser nhận diện nếu có `@Injectable`.
- Nếu provider bị disable xuất hiện trong chain dependency của `@Root`, quá trình generate sẽ dừng và báo lỗi.
- Nếu provider bị disable không được dùng trong graph hiện tại, Dix sẽ bỏ qua provider đó trong runtime wiring flow.

## Lỗi thường gặp

Xem danh sách lỗi và cách khắc phục tại: [Lỗi thường gặp](/docs/common-errors).
