FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o /out/user-service ./cmd/main.go

FROM alpine:3.20

RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app

COPY --from=builder /out/user-service /app/user-service

EXPOSE 50051

ENTRYPOINT ["/app/user-service"]


