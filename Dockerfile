# ---- Stage 1: Build ----
# Sử dụng image Go chính thức làm base image cho việc build
# Chọn phiên bản Go phù hợp với go.mod của bạn (ví dụ: 1.21, dựa trên go.mod là 1.24.2)
FROM golang:1.24.2-alpine AS builder

# Thiết lập thư mục làm việc bên trong container
WORKDIR /app

# Sao chép toàn bộ mã nguồn của ứng dụng
COPY . .

# Sao chép file go.mod và go.sum để tải dependencies
# Tận dụng cache của Docker: chỉ khi các file này thay đổi thì mới tải lại dependencies
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify


# Build ứng dụng Go.
# Entry point của bạn là cmd/server/main.go
# Output sẽ là một file thực thi tên là 'gochat-backend' trong thư mục /app
RUN go build -ldflags="-w -s" -o /app/gochat-backend ./cmd/server/main.go


# ---- Stage 2: Run ----
FROM alpine:latest AS runner
WORKDIR /app

COPY --from=builder /app/gochat-backend .
COPY .env .


# Khai báo cổng mà ứng dụng sẽ lắng nghe (metadata, không thực sự mở cổng)
# Dựa trên config/config.go, PORT mặc định là 8080
EXPOSE 8080

# Lệnh để chạy ứng dụng khi container khởi động
# Đảm bảo đường dẫn đến binary là chính xác
CMD ["/app/gochat-backend"]
