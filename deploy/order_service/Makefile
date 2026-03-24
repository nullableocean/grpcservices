.PHONY: up down logs clean

COMPOSE_FILE=compose.yml

up:
	docker compose -f ${COMPOSE_FILE} up -d --build

down:
	docker compose -f ${COMPOSE_FILE} down

restart:
	docker compose -f ${COMPOSE_FILE} down order_service && docker compose -f ${COMPOSE_FILE} up -d --build order_service

restart-all: down up

logs:
	docker compose -f ${COMPOSE_FILE} logs

clean:
	docker compose -f ${COMPOSE_FILE} down -v