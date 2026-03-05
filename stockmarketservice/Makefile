.PHONY: up down logs restart
 
up:
	docker compose -f compose.dev.yml up -d --build

down:
	docker compose -f compose.dev.yml down

restart: down up

logs:
	docker compose -f compose.dev.yml logs