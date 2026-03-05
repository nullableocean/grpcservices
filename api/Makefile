.PHONY: gen

MODULE = github.com/nullableocean/grpcservices/api

gen: 
	protoc -I=proto/ \
		--go_opt=module=$(MODULE) \
		--go_out=. \
		--go-grpc_opt=module=$(MODULE) \
		--go-grpc_out=. \
		./proto/types/*.proto ./proto/service/*.proto ./proto/events/order/*.proto