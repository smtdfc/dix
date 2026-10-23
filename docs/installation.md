# Cài đặt Dix

## Yêu cầu

- Go 1.21+ (dùng generics cho `di.Singleton[T]`)

## Cài đặt

```bash
go install github.com/smtdfc/dix@latest
```

## Kiểm tra

```bash
dix --help
```

Nếu thấy danh sách lệnh `run`, `build`, `wire`, `cache` là đã cài thành công.

## Cập nhật

```bash
go install github.com/smtdfc/dix@latest
```

## Xây dựng từ source

```bash
git clone https://github.com/smtdfc/dix.git
cd dix
go build -o dix .
./dix --help
```
