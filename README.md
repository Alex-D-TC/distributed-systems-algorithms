Generate the protobuf files

protoc --go_out=plugins=grpc:. ./protocol/*.proto

Generate the docker image

docker build -t distributed-systems

Spin up a testnet

Go to the tesnet folder and run:

docker-compose up

