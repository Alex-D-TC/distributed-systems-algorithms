FROM ubuntu:18.04

WORKDIR /cmnd

COPY ./protocol/commands.proto commands.proto

RUN apt update
RUN apt install wget -y
RUN wget https://github.com/fullstorydev/grpcurl/releases/download/v1.2.1/grpcurl_1.2.1_linux_x86_64.tar.gz
RUN tar -xzf grpcurl_1.2.1_linux_x86_64.tar.gz

CMD tail -f /dev/null
