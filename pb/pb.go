package pb

//go:generate protoc -I . --go_out=plugins=grpc:. api.proto models.proto
//go:generate protoc -I . --js_out=import_style=commonjs:../web/src/pb --grpc-web_out=import_style=commonjs,mode=grpcwebtext:../web/src/pb api.proto models.proto

