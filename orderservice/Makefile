.PHONY: test up down logs restart

SERVICE_NAME=orderapp

test:
	go test -v ./internal/service/order
	go test -v ./tests

up:
	docker compose -f compose.dev.yml up -d --build

down:
	docker compose -f compose.dev.yml down

restart:
	docker compose -f compose.dev.yml down ${SERVICE_NAME}
	docker compose -f compose.dev.yml up ${SERVICE_NAME} -d --build

logs:
	docker compose -f compose.dev.yml logs orderapp -f