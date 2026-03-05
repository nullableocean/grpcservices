.PHONY: test up down logs

test:
	go test -v ./internal/service/spot
	go test -v ./tests

up:
	docker compose -f compose.dev.yml up -d --build

down:
	docker compose -f compose.dev.yml down

logs:
	docker compose -f compose.dev.yml logs