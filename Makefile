
proto:
	protoc -I ./internal/proto internal/proto/dgrep/dgrep.proto --go_out=./internal/gen/go/ --go_opt=paths=source_relative  --go-grpc_out=./internal/gen/go/ --go-grpc_opt=paths=source_relative