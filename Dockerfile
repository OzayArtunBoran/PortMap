# ── Stage 1: Build ──
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o /app/bin/portmap .

# ── Stage 2: Runtime ──
FROM alpine:3.19
RUN apk add --no-cache ca-certificates procps lsof
COPY --from=builder /app/bin/portmap /usr/local/bin/
ENTRYPOINT ["portmap"]
