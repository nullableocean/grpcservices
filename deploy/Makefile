SPOT_DIR = spot_instrument
ORDER_DIR = order_service
METRICS_DIR = metrics
APP_NETWORK = appnet

.PHONY: help up up-networks up-spot up-order up-metrics down clean

help:
	@echo "Available targets:"
	@echo "  up-networks   Create networks (run first)"
	@echo "  up            Start all services (requires networks)"
	@echo "  up-spot       Start spot_instrument + its DB"
	@echo "  up-order      Start order_service + its DB"
	@echo "  up-metrics    Start monitoring stack"
	@echo "  down          Stop all services"
	@echo "  clean         Stop all services and remove networks"

up: create-net
	make up-metrics
	@sleep 5
	make up-spot
	@sleep 3
	make up-order

	@echo "grafana: http://localhost:3000"
	@echo "jaeger: http://localhost:16686"

up-srvs: up-spot up-order

up-spot:
	cd $(SPOT_DIR) && make up

up-order:
	cd $(ORDER_DIR) && make up

up-metrics:
	cd $(METRICS_DIR) && make up

down-spot:
	cd $(SPOT_DIR) && make down

down-order:
	cd $(ORDER_DIR) && make down

down-metrics:
	cd $(METRICS_DIR) && make down

down-srvs: down-spot down-order

restart-srvs:
	cd $(SPOT_DIR) && make restart
	cd $(ORDER_DIR) && make restart

clean-srvs:
	cd $(SPOT_DIR) && make clean
	cd $(ORDER_DIR) && make clean

down: down-spot down-order down-metrics rm-net

create-net:
	docker network inspect ${APP_NETWORK} >/dev/null 2>&1 || docker network create ${APP_NETWORK}

rm-net:
	-docker network rm ${APP_NETWORK} 2>/dev/null