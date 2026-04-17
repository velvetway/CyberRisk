FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S app -G app
COPY --from=builder /out/server /usr/local/bin/server
USER app
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/server"]
