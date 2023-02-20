FROM golang:1.19-alpine AS builder
WORKDIR /usr/src/iploc
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o iploc ./cmd/app

FROM alpine:3.17.0
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
COPY --from=builder /usr/src/iploc/iploc /usr/local/bin/iploc
WORKDIR /tmp
CMD ["iploc"]