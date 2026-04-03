.PHONY: up down logs

up:
	docker compose -f compose.yml up --build -d

down:
	docker compose -f compose.yml down

restart:
	docker compose -f compose.yml restart

logs:
	docker compose -f compose.yml logs

clean:
	docker compose -f compose.yml down -v