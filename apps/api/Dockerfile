FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/server/
RUN CGO_ENABLED=0 GOOS=linux go build -o /migrate ./cmd/migrate/

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app /app
COPY --from=builder /migrate /usr/local/bin/migrate

EXPOSE 8080
ENTRYPOINT ["/app"]
