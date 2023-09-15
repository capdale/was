FROM ubuntu:22.04

LABEL author="devhoodit"

RUN apt-get update && \
    apt-get -y install sudo && \
    sudo apt-get -y install systemctl && \
    sudo apt-get -y install wget

# install services
RUN wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz && \
    sudo rm -rf /usr/local/go && tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

ENV PATH "/usr/local/go/bin:${PATH}"

WORKDIR /server

COPY go.mod go.sum ./

RUN go mod download

COPY ./ ./

EXPOSE 8080

CMD [ "/bin/bash", "-c", "go build main.go" ]

ENTRYPOINT [ "/bin/bash", "-c", "go run main.go" ]