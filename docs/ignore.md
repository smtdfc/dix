# Bỏ qua package (Ignore Packages)

Khi dự án lớn, bạn có thể không muốn Dix quét các package như `internal/mock`, mã cũ, hoặc thư viện bên thứ ba.

## Cấu hình qua `dix.config.json`

```json
{
  "ignore_packages": [
    "internal/mock",
    "github.com/foo/bar/...",
    "github.com/baz/qux/*"
  ],
  "output": "./generated/dix/root.go"
}
```

## Cấu hình qua CLI

Flag `--ignore-package` có thể lặp nhiều lần và được ghép với giá trị trong `dix.config.json`:

```bash
dix run . --ignore-package internal/mock --ignore-package github.com/foo/bar
dix build . --ignore-package github.com/foo/bar/...
dix wire --workspace . --ignore-package internal/mock
```

## Các mẫu pattern hỗ trợ

| Mẫu | Khớp với |
|-----|----------|
| `github.com/foo/bar` | Chính xác gói `bar` (không khớp sub-package) |
| `github.com/foo/bar/...` | Gói `bar` và mọi gói con bên dưới |
| `github.com/foo/bar/*` | Chính xác `bar` và gói con trực tiếp `bar/*` |
| `github.com/foo/*bar*` | Glob, `*` đại diện cho mọi ký tự trừ `/` |
| `internal/mock` | Chính xác `internal/mock` (không false-positive) |

> Pattern thuần túy là exact match, nên `"foo"` sẽ **không** khớp `"foobar"`.

## Kết quả

Khi có ít nhất một package bị bỏ qua, Dix in tóm tát thay vì spam từng dòng:

```
[Ignore] 3 package(s) ignored, 42 scanned
```

## Cách chọn pattern

- Muốn bỏ qua một gói con duy nhất: dùng exact `github.com/foo/bar`.
- Muốn bỏ qua cả gói và mọi sub-package: dùng `github.com/foo/bar/...`.
- Muốn bỏ qua gói cha và cấp con trực tiếp: dùng `github.com/foo/bar/*`.
- Muốn bỏ qua theo mẫu tên: dùng glob `github.com/foo/*bar*`.
