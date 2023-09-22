FROM ubuntu:22.04

LABEL author="devhoodit"

RUN apt-get update && \
    apt-get -y install sudo && \
    sudo apt-get -y install systemctl && \
    sudo apt-get -y install wget

WORKDIR /server

COPY ./was-linux-amd64 ./
COPY ./config.yaml ./

EXPOSE 8080

ENTRYPOINT [ "./was-linux-amd64" ]