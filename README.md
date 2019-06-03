Generate the protobuf files

protoc --go_out=plugins=grpc:. ./protocol/*.proto

If the command fails, install the protoc-gen-go package with the following command
go get -u github.com/golang/protobuf/protoc-gen-go

Generate the docker image

docker build -t distributed-systems

Spin up a testnet

Go to the tesnet folder and run:

docker-compose up
