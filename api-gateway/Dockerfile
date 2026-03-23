FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o c2graph-api ./cmd/server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/c2graph-api /usr/local/bin/c2graph-api
ENTRYPOINT ["c2graph-api"]
