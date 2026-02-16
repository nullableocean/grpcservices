genapi:
	protoc -I api --go_opt=module=main --go_out=. \
	--go-grpc_opt=module=main --go-grpc_out=. \
	api/*.proto