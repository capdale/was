FROM ubuntu:22.04

LABEL author="devhoodit"

RUN apt-get update && \
    apt-get -y install sudo && \
    sudo apt-get -y install systemctl && \
    sudo apt-get -y install wget

WORKDIR /server

COPY ./was-linux-amd64 ./
COPY ./config.yaml ./

ENV GIN_MODE=release

EXPOSE 443

ENTRYPOINT [ "./was-linux-amd64" ]