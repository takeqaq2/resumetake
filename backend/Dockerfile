FROM golang:1.25-alpine AS builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o server .

FROM alpine:3.21
RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.tuna.tsinghua.edu.cn#g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata wget tesseract-ocr poppler-utils su-exec
WORKDIR /app
RUN addgroup -S app && adduser -S -G app -u 10001 app
COPY --from=builder /app/server .
RUN mkdir -p /app/data && chown -R app:app /app
EXPOSE 8000
ENV TZ=Asia/Shanghai
HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -qO- http://localhost:8000/api/health || exit 1
# Entrypoint runs as root to fix the data-volume ownership (named volumes may be
# owned by root from a prior root-run image), then drops privileges to the app user.
ENTRYPOINT ["sh", "-c", "chown -R app:app /app/data && exec su-exec app ./server"]
