FROM golang:1.21-alpine AS builder

ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

WORKDIR /build

RUN apk add --no-cache git make gcc musl-dev sqlite-dev

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o nta-server ./cmd/nta-server

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata sqlite

RUN addgroup -g 1000 nta && \
    adduser -D -u 1000 -G nta nta

WORKDIR /app

COPY --from=builder /build/nta-server /app/
COPY --from=builder /build/config /app/config
COPY --from=builder /build/zeek-scripts /app/zeek-scripts

RUN mkdir -p /app/data /app/logs /app/reports && \
    chown -R nta:nta /app

USER nta

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/nta-server"]
CMD ["-config", "/app/config/nta.yaml"]