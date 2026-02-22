PACKAGE="github.com/nullableocean/grpcservices"

.PHONY: up down up-monitoring localbuild-spot localbuild-spot up-mon up-srvs test

localbuild-spot:
	cd spot && go build -o bin/spot ./cmd
	cp spot/.env spot/bin/.env

localbuild-order:
	cd order && go build -o bin/orderserv ./cmd/server
	cp order/.env order/bin/.env

localbuild-order-cli:
	cd order && go build -o bin/ordercli ./cmd/client

# up all monintoring + services
up: genapi tidy mod-download up-mon up-srvs
	@echo "=== OK ==="
	@echo "grafana: http://localhost:3000"
	@echo "jaeger: http://localhost:16686"

down: down-srvs down-mon
	@echo "=== SERVICES DOWNED === "

#restart + rebuild compose only services
restart-srvs: down-srvs up-srvs

up-srvs:
	@echo "=== UP SPOT SERVICE ==="
	docker compose -f spot/compose.dev.yml up -d --build
	@echo "=== UP ORDER SERVICE ==="
	docker compose -f order/compose.dev.yml up -d --build

up-mon:
	@echo "=== UP METRICS PANELS ==="
	docker compose -f metrics/compose.yml up -d
	@sleep 3

down-srvs:
	@echo "=== DOWN ORDER SERVICE ==="
	docker compose -f order/compose.dev.yml down
	@echo "=== DOWN SPOT SERVICE ==="
	docker compose -f spot/compose.dev.yml down

down-mon:
	@echo "=== DOWN METRICS PANELS ==="
	docker compose -f metrics/compose.yml down
	@sleep 3


logs-order:
	docker compose -f order/compose.dev.yml logs orderapp

logs-spot:
	docker compose -f spot/compose.dev.yml logs spotapp


test:
	@echo "\n\nSPOT SERVICE TESTS\n"
	cd spot && make test
	@echo "\n\nORDER SERVICE TESTS\n"
	cd order && make test

tidy:
	cd pkg   && go mod tidy
	cd api   && go mod tidy
	cd spot  && go mod tidy
	cd order && go mod tidy

mod-download:
	cd pkg   && go mod download
	cd api   && go mod download
	cd spot  && go mod download
	cd order && go mod download

genapi:
	protoc -I api --go_opt=module=${PACKAGE} --go_out=. \
	--go-grpc_opt=module=${PACKAGE} --go-grpc_out=. \
	api/*.proto
	cd api && go mod tidy