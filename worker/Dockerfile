FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o c2graph-worker ./cmd/worker/
RUN CGO_ENABLED=0 go build -o c2graph-rescore ./cmd/rescore/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/c2graph-worker /usr/local/bin/c2graph-worker
COPY --from=builder /app/c2graph-rescore /usr/local/bin/c2graph-rescore
ENTRYPOINT ["c2graph-worker"]
