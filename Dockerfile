FROM golang:1.19-alpine

RUN apk add --no-cache tzdata

ENV TZ=Europe/Moscow

WORKDIR /usr/src/iploc/

COPY . .

RUN cd ./cmd/app/ && go build -v -o /usr/local/bin/iploc .

WORKDIR /tmp

CMD ["iploc"]