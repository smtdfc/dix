# Scan Cache

Dix lưu metadata đã quét thành công vào `.dix/cache/<fingerprint>.json`.

## Fingerprint

Fingerprint được tính từ:
- Phiên bản Go (`runtime.Version()`)
- Chế độ scan (`--workspace` hay không)
- Nội dung các file `.go`, `go.mod`, `go.sum` trong thư mục quét
- File `go.work` nếu dùng `--workspace`

Cache bị hủy và quét lại khi bất kỳ yếu tố nào trên thay đổi.

## Lệnh liên quan

| Lệnh | Ý nghĩa |
|------|---------|
| `dix run --no-cache .` | Ép quét lại, không dùng cache |
| `dix build --no-cache` | Ép quét lại, không dùng cache |
| `dix wire --no-cache` | Ép quét lại, không dùng cache |
| `dix cache clean .` | Xóa toàn bộ cache tại `.dix/cache` |

## Vị trí cache

Cache nằm trong thư mục gốc dự án:

```
<project>/.dix/cache/<fingerprint>.json
```

Tệp metadata intermediate (`scan_<timestamp>.dix`) được lưu tại thư mục hiện hành khi chạy `run`/`build`/`wire`.
