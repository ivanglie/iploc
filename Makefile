.PHONY: run tests docker-dev

run:
	go run -race ./cmd/app/ --local --dbg

tests:
	go test -v -cover -race ./...

docker-dev:
	docker compose -f docker-compose.dev.yml down -v && docker compose -f docker-compose.dev.yml up --build -d