package proto

//go:generate protoc -I ./ --go_out=plugins=grpc:./chat ./chat.proto
//go:generate protoc -I ./ --go_out=plugins=grpc:./greeter ./greeter.proto
