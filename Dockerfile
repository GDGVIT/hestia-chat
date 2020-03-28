# Builder
FROM golang:alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

ENV GO111MODULE=on

RUN CGO_ENABLED=0 GOOS=linux go build -o ./bin/main ./server.go


# Runner
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/bin/main .

CMD ["./main"]