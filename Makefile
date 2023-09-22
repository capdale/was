APP_NAME=was
BUILD_DIR=./build/
DOCKER_DIR=./docker/

PATH_RESOLVE=MSYS2_ARG_CONV_EXCL='*'
DOCKER_RUN=docker run -v .:/src -w /src
DOCKER_IMAGE=golang:latest

cp-config:
	cp config.yaml ./docker/config.yaml

# linux
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_DIR}${APP_NAME}-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o ${BUILD_DIR}${APP_NAME}-linux-arm64

# build-linux-riscv64:
# 	GOOS=linux GOARCH=riscv64 go build -o ${BUILD_DIR}${APP_NAME}-linux-riscv64

# window
# build-windows-386:
# 	GOOS=windows GOARCH=386 go build -o ${BUILD_DIR}${APP_NAME}-windows-386.exe

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o ${BUILD_DIR}${APP_NAME}-windows-amd64.exe

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -o ${BUILD_DIR}${APP_NAME}-windows-arm64.exe

# build all
build: build-linux-amd64 build-linux-arm64 build-windows-amd64 build-windows-arm64

# docker
docker-linux-amd64: cp-config
	cp ${BUILD_DIR}${APP_NAME}-linux-amd64 ${DOCKER_DIR}
	docker build --tag was ${DOCKER_DIR}.

docker-linux-arm64: cp-config
	cp ${BUILD_DIR}${APP_NAME}-linux-arm64 ${DOCKER_DIR}
	docker build --tag was -f ${DOCKER_DIR}Dockerfile.arm64

# default
docker-build:
	${PATH_RESOLVE} ${DOCKER_RUN} -e GOOS=linux -e GOARCH=amd64 ${DOCKER_IMAGE} make build-linux-amd64