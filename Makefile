TOKEN=$(shell cat ip2location.secret)

tests:
	go test -v -cover -race ./...

run:
	go run -race ./cmd/app/ --token=$(TOKEN)

dc:
	TOKEN=$(TOKEN) docker compose up -d