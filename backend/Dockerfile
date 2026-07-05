FROM golang:1.22-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8000
ENV TZ=Asia/Shanghai
CMD ["./server"]