module github.com/nullableocean/grpcservices/orderserviceclient

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	google.golang.org/grpc v1.79.1
)

require (
	github.com/google/uuid v1.6.0
	github.com/nullableocean/grpcservices/api v0.0.0
	github.com/nullableocean/grpcservices/shared v0.0.0
	github.com/shopspring/decimal v1.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/nullableocean/grpcservices/api => ../api
	github.com/nullableocean/grpcservices/shared => ../shared
)
