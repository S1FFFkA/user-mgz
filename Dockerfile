FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /out/user-service ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /out/push-consumer ./cmd/push-consumer/main.go

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app

COPY --from=builder /out/user-service /app/user-service
COPY --from=builder /out/push-consumer /app/push-consumer

EXPOSE 50051

ENTRYPOINT ["/app/user-service"]


