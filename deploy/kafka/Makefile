.PHONY: up down clean logs

COMPOSE_FILE=compose.yml

up:
	docker compose -f ${COMPOSE_FILE} up -d --build --wait

down:
	docker compose -f ${COMPOSE_FILE} down

logs:
	docker compose -f ${COMPOSE_FILE} logs

clean:
	docker compose -f ${COMPOSE_FILE} down -v