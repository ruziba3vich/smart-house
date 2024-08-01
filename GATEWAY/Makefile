generate-rpc:
	protoc \
	--go_out=genprotos \
	--go_opt=paths=source_relative \
	--go-grpc_out=genprotos \
	--go-grpc_opt=paths=source_relative \
	submodules/users_submodule/protos/users.proto

generate-devices-rpc:
	protoc \
	--go_out=genprotos \
	--go_opt=paths=source_relative \
	--go-grpc_out=genprotos \
	--go-grpc_opt=paths=source_relative \
	devices_submodule/devices.proto

generate-control-rpc:
	protoc \
	--go_out=genprotos \
	--go_opt=paths=source_relative \
	--go-grpc_out=genprotos \
	--go-grpc_opt=paths=source_relative \
	controller_submodule/controller.proto

swagg:
	swag init -g app/handler/handler.go
