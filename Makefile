.PHONY: up down up-monitoring localbuild-spot localbuild-spot up-mon up-srvs test logs create-net rm-net up-broker down-broker

MONITORING_NET=monitoringnet
BROKER_NET=brokernet
SERVICES_NET=grpcservices


genapi:
	cd api && make gen

# up all monintoring + services
up: genapi create-net up-mon up-broker up-srvs
	@echo "=== OK ==="
	@echo "grafana: http://localhost:3000"
	@echo "jaeger: http://localhost:16686"

down: down-srvs down-mon down-broker
	@echo "=== SERVICES DOWNED === "


create-net:
	docker network inspect ${MONITORING_NET} >/dev/null 2>&1 || docker network create ${MONITORING_NET}
	docker network inspect ${BROKER_NET} >/dev/null 2>&1 || docker network create ${BROKER_NET}
	docker network inspect ${SERVICES_NET} >/dev/null 2>&1 || docker network create ${SERVICES_NET}

rm-net:
	-docker network rm ${MONITORING_NET} 2>/dev/null
	-docker network rm ${BROKER_NET} 2>/dev/null
	-docker network rm ${SERVICES_NET} 2>/dev/null



up-mon:
	@echo "=== UP METRICS PANELS ==="
	cd metrics && make up
	@sleep 3

down-mon:
	@echo "=== DOWN METRICS PANELS ==="
	docker compose -f metrics/compose.yml down

up-broker:
	@echo "=== UP KAFKA BROKER ==="
	cd kafka && make up
	@sleep 3

down-broker:
	@echo "=== DOWN KAFKA BROKER ==="
	cd kafka && make down

up-srvs:
	@echo "=== UP USER SERVICE ==="
	cd userservice && make up

	@echo "=== UP STOCKMARKET SERVICE ==="
	cd stockmarketservice && make up

	@echo "=== UP SPOT SERVICE ==="
	cd spotinstrument && make up

	@echo "=== UP ORDER SERVICE ==="
	cd orderservice && make up

restart-srvs: down-srvs up-srvs

down-srvs:
	@echo "=== DOWN ORDER SERVICE ==="
	cd orderservice && make down

	@echo "=== DOWN STOCKMARKET SERVICE ==="
	cd stockmarketservice && make down

	@echo "=== DOWN SPOT SERVICE ==="
	cd spotinstrument && make down

	@echo "=== DOWN USER SERVICE ==="
	cd userservice && make down


logs:
	@echo "=== LOGS ORDER SERVICE ==="
	cd orderservice && make logs
	@echo
	@echo
	@echo "=== LOGS STOCKMARKET SERVICE ==="
	cd stockmarketservice && make logs
	@echo
	@echo
	@echo "=== LOGS SPOT SERVICE ==="
	cd spotinstrument && make logs
	@echo
	@echo
	@echo "=== LOGS USER SERVICE ==="
	cd userservice && make logs

tidy:
	cd shared   && go mod tidy
	cd api   && go mod tidy
	cd spotinstrument  && go mod tidy
	cd orderservice && go mod tidy
	cd stockmarketservice && go mod tidy
	cd userservice && go mod tidy

mod-download:
	cd shared   && go mod download
	cd api   && go mod download
	cd spotinstrument  && go mod download
	cd orderservice && go mod download
	cd stockmarketservice && go mod download
	cd userservice && go mod download

