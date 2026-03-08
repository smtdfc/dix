# Singleton trong Dix

## Tổng quan

Trong Dix, `Singleton` là cách khai báo dependency theo wrapper type ở tham số hàm, dùng kiểu:

- `di.Singleton[T]`

Mục tiêu của wrapper này là giúp bạn biểu diễn rõ ràng dependency cần được truyền theo ngữ nghĩa singleton trong graph.

## Cấu trúc `Singleton[T]`

Trong package `di`, cấu trúc hiện tại là:

```go
type Singleton[T any] struct {
	Instance T
}
```

Khởi tạo wrapper:

```go
s := di.NewSingleton(value)
```

Trong đó `value` có kiểu `T`.

## Cách unwrap để lấy giá trị

Bạn có 2 cách lấy lại giá trị bên trong wrapper:

1. Dùng method `Get()` (khuyến nghị)
2. Truy cập trực tiếp field `Instance`

Ví dụ:

```go
func UseRepo(repoS di.Singleton[*Repo]) {
	repo1 := repoS.Get()       // cach khuyen nghi
	repo2 := repoS.Instance    // cach truc tiep
	_, _ = repo1, repo2
}
```

Khuyến nghị dùng `Get()` để code nhất quán và dễ thay đổi implementation sau này.

## Quy tắc khai báo

1. Chỉ dùng đúng dạng `di.Singleton[T]`.
2. Không dùng con trỏ wrapper: `*di.Singleton[T]` là không hợp lệ.
3. `T` phải khớp với provider trả về trong graph.

## Ví dụ đúng

```go
package app

import "github.com/smtdfc/dix/di"

type Repo struct{}
type App struct{}

// @Injectable
func NewRepo() *Repo {
	return &Repo{}
}

// @Injectable
// @Root
func NewApp(repo di.Singleton[*Repo]) *App {
	_ = repo
	return &App{}
}
```

## Ví dụ sai

```go
package app

import "github.com/smtdfc/dix/di"

type Repo struct{}
type App struct{}

// @Injectable
func NewRepo() *Repo {
	return &Repo{}
}

// @Injectable
// @Root
func NewApp(repo *di.Singleton[*Repo]) *App {
	return &App{}
}
```

Lỗi parser tương ứng:

`parser/validation: singleton dependency must be di.Singleton[T], not *di.Singleton[T]`

## Hành vi generate code

Khi gặp singleton dependency, Dix sẽ tạo object mới và bọc trực tiếp bằng `di.NewSingleton(...)` tại call-site.

Ví dụ dạng generate:

```go
repo0 := pkg.NewRepo()
app0 := pkg.NewApp(di.NewSingleton(pkg.NewRepo()))
```

Điểm cần lưu ý:

- Dependency thường: lấy từ container scope đã generate.
- Dependency singleton: bọc trực tiếp bằng `di.NewSingleton(...)` khi truyền vào provider cần singleton.

Nếu cần unwrap ngay trong provider, bạn có thể làm như sau:

```go
func NewApp(repoS di.Singleton[*Repo]) *App {
	repo := repoS.Get()
	_ = repo
	return &App{}
}
```

## Lỗi thường gặp

Xem danh sách lỗi và cách khắc phục tại: [Lỗi thường gặp](/docs/common-errors).
