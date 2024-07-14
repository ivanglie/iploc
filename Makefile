TOKEN=$(shell cat ip2location.secret)

tests:
	go test -v -cover -race ./...

run:
	go run -race ./cmd/app/ --token=$(TOKEN)

debug:
	go run -race ./cmd/app/ --token=$(TOKEN) --dbg

local:
	go run -race ./cmd/app/ --token=$(TOKEN) --local --dbg

dc:
	TOKEN=$(TOKEN) docker compose up -d