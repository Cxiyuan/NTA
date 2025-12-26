FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make gcc musl-dev

COPY go.mod ./
RUN go mod download && go mod tidy

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -v \
    -ldflags="-w -s" \
    -o kafka-consumer ./cmd/kafka-consumer

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/kafka-consumer /app/

RUN addgroup -g 1000 nta && \
    adduser -D -u 1000 -G nta nta && \
    chown -R nta:nta /app

USER nta

ENTRYPOINT ["/app/kafka-consumer"]