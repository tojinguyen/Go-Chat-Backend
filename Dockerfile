# ---- Stage 1: Build ----
# Sử dụng image Go chính thức làm base image cho việc build
# Chọn phiên bản Go phù hợp với go.mod của bạn (ví dụ: 1.21, dựa trên go.mod là 1.24.2)
FROM golang:1.24.2-alpine AS builder

# Thiết lập biến môi trường cho build, đặc biệt quan trọng với Go
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Thiết lập thư mục làm việc bên trong container
WORKDIR /app

# Sao chép file go.mod và go.sum để tải dependencies
# Tận dụng cache của Docker: chỉ khi các file này thay đổi thì mới tải lại dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Sao chép toàn bộ mã nguồn của ứng dụng
COPY . .

# Build ứng dụng Go.
# Entry point của bạn là cmd/server/main.go
# Output sẽ là một file thực thi tên là 'gochat-backend' trong thư mục /app
RUN go build -ldflags="-w -s" -o /app/gochat-backend ./cmd/server/main.go

# ---- Stage 2: Final Image ----
# Sử dụng một image cơ sở rất nhẹ. Alpine Linux là một lựa chọn tốt.
FROM alpine:latest

# (Tùy chọn) Cài đặt các certificates cần thiết nếu app của bạn gọi HTTPS ra bên ngoài,
# hoặc các dependencies hệ thống khác nếu binary của bạn không phải static hoàn toàn.
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Sao chép file binary đã build từ stage 'builder'
COPY --from=builder /app/gochat-backend /app/gochat-backend

# Sao chép thư mục migrations và docs (nếu bạn cần chúng trong container)
# Ví dụ: nếu bạn chạy migration từ bên trong container hoặc serve swagger docs
COPY migrations ./migrations
COPY docs ./docs

# Khai báo cổng mà ứng dụng sẽ lắng nghe (metadata, không thực sự mở cổng)
# Dựa trên config/config.go, PORT mặc định là 8080
EXPOSE 8080

# Lệnh để chạy ứng dụng khi container khởi động
# Đảm bảo đường dẫn đến binary là chính xác
CMD ["/app/gochat-backend"]

COPY .env .env
