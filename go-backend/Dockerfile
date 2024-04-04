# 使用官方Go镜像作为构建环境
FROM golang:1.21 as builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum并下载依赖项
COPY go.* ./
RUN go mod download

# 复制源代码到容器中
COPY . .

# 编译Go应用为一个二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app

# 使用scratch作为最终镜像，减小镜像体积
FROM scratch
COPY --from=builder /app/app ./
CMD ["./app"]
