PACKAGE="github.com/nullableocean/grpcservices"

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