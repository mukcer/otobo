# Билд стадии
FROM golang:1.25-alpine AS builder
EXPOSE 3000
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
#RUN go build -o otobo ./cmd/app
RUN CGO_ENABLED=0 go build -gcflags="all=-N -l" -o otobo ./cmd/app

#RUN go install github.com/go-delve/delve/cmd/dlv@latest
# Финальная стадия
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/otobo .
#COPY --from=builder /go/bin/dlv /
# Запускаем приложение под управлением Delve
#CMD ["/dlv", "--listen=:40000", "--headless=true", "--log", "--accept-multiclient", "--api-version=2", "exec", "./otobo"]

CMD ["./otobo"]
