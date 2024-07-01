FROM golang:1.19-alpine AS builder
WORKDIR /usr/src/iploc
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -v -o iploc cmd/app/main.go

FROM --platform=$BUILDPLATFORM alpine:3.17.0 
WORKDIR /usr/local/bin/
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow
COPY --from=builder /usr/src/iploc/iploc /usr/local/bin/iploc
COPY --from=builder /usr/src/iploc/web/template /usr/local/bin/web/template
CMD ["iploc"]