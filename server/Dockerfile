# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -trimpath -ldflags="-s -w" -o /out/gin-react-admin .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
COPY --from=builder /out/gin-react-admin ./gin-react-admin
COPY --from=builder /src/database ./database

RUN mkdir -p runtime/uploads \
    && chown -R app:app /app

USER app
ENV SERVER_ADDR=:8080
EXPOSE 8080

VOLUME ["/app/runtime"]
ENTRYPOINT ["/app/gin-react-admin"]
