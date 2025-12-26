FROM golang:1.23-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make gcc musl-dev

COPY . .

RUN go mod download && go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o notification-service ./services/notification-service/cmd

FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/notification-service .

EXPOSE 8086

ENTRYPOINT ["/app/notification-service"]
