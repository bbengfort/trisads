package proto

//go:generate protoc -I . --go_out=plugins=grpc:. api.proto models.proto
