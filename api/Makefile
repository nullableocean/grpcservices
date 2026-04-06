.PHONY: gen install-tools

MODULE = github.com/nullableocean/grpcservices/api

PROTOC_GEN_GO_VERSION = v1.36.11
PROTOC_GEN_GO_GRPC_VERSION = latest
PROTOC_GEN_VALIDATE_VERSION = v1.3.3

install-tools:
	@echo "Installing protoc-gen-go..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@echo "Installing protoc-gen-go-grpc..."
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@echo "Installing protoc-gen-validate..."
	go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)

copy-validate:
	@echo "Copying validate.proto to proto/validate/"
	@mkdir -p proto/validate
	@go mod download github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@cp $(shell go list -f '{{ .Dir }}' -m github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION))/validate/validate.proto proto/validate/

gen:
	protoc -I=proto/ \
		--go_opt=module=$(MODULE) \
		--go_out=. \
		--go-grpc_opt=module=$(MODULE) \
		--go-grpc_out=. \
		--validate_opt=module=$(MODULE) \
		--validate_out="lang=go:." \
		./proto/models/*.proto \
		./proto/service/*.proto \
		./proto/events/order/*.proto \
		./proto/events/markets/*.proto