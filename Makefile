PACKAGE="github.com/nullableocean/grpcservices"

.PHONY: up down up-monitoring localbuild-spot localbuild-spot

localbuild-spot:
	cd spot && go build -o bin/spot ./cmd
	cp spot/.env spot/bin/.env

localbuild-order:
	cd order && go build -o bin/order ./cmd
	cp order/.env order/bin/.env

up: genapi tidy mod-download
	@echo "=== UP METRICS PANELS ==="
	docker compose -f metrics/compose.yml up -d
	@sleep 3
	@echo "=== UP SPOT SERVICE ==="
	docker compose -f spot/compose.dev.yml up -d --build
	@echo "=== UP ORDER SERVICE ==="
	docker compose -f order/compose.dev.yml up -d --build
	@echo "=== OK ==="
	@echo "grafana: http://localhost:3000"

mod-download:
	cd pkg   && go mod download
	cd api   && go mod download
	cd spot  && go mod download
	cd order && go mod download

down:
	docker compose -f order/compose.dev.yml down
	docker compose -f spot/compose.dev.yml down
	docker compose -f metrics/compose.yml down
	@echo "=== SERVICES DOWNED === "


tidy:
	cd pkg   && go mod tidy
	cd api   && go mod tidy
	cd spot  && go mod tidy
	cd order && go mod tidy

genapi:
	protoc -I api --go_opt=module=${PACKAGE} --go_out=. \
	--go-grpc_opt=module=${PACKAGE} --go-grpc_out=. \
	api/*.proto
	cd api && go mod tidy