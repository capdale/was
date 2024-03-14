APP_NAME=was
BUILD_DIR=./build/
DOCKER_DIR=./docker/

PATH_RESOLVE=MSYS2_ARG_CONV_EXCL='*'
DOCKER_RUN=docker run -v .:/src/app -w /src/app
DOCKER_IMAGE=golang:latest



# linux
build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -buildvsc=false -o ${BUILD_DIR}${APP_NAME}-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o ${BUILD_DIR}${APP_NAME}-linux-arm64

# build-linux-riscv64:
# 	GOOS=linux GOARCH=riscv64 go build -o ${BUILD_DIR}${APP_NAME}-linux-riscv64

# window
# build-windows-386:
# 	GOOS=windows GOARCH=386 go build -o ${BUILD_DIR}${APP_NAME}-windows-386.exe

build-windows-amd64:
	GOOS=windows GOARCH=amd64 go build -o ${BUILD_DIR}${APP_NAME}-windows-amd64.exe

# build-windows-arm64:
# 	GOOS=windows GOARCH=arm64 go build -o ${BUILD_DIR}${APP_NAME}-windows-arm64.exe

# build all
build-all: build-linux-amd64 build-linux-arm64 build-windows-amd64

# default build
build: build-linux-amd64


# build in docker
# linux
docker-build-linux-amd64:
	${DOCKER_RUN} ${DOCKER_IMAGE} make build-linux-amd64

docker-build-linux-arm64:
	${DOCKER_RUN} ${DOCKER_IMAGE} make build-linux-arm64

# window
docker-build-windows-amd64:
	${DOCKER_RUN} ${DOCKER_IMAGE} make build-widnows-amd64

# docker-build-windows-arm64:
# 	${DOCKER_RUN} ${DOCKER_IMAGE} make build-widnows-arm64

# docker build all
docker-build-all: docker-build-linux-amd64 docker-build-linux-arm64 docker-build-windows-amd64

# default
docker-build: docker-build-linux-amd64

# build docker image
docker-image-linux-amd64:
	cp ${BUILD_DIR}${APP_NAME}-linux-amd64 ${DOCKER_DIR}
	docker build --tag ${APP_NAME} ${DOCKER_DIR}.

docker-image-linux-arm64:
	cp ${BUILD_DIR}${APP_NAME}-linux-arm64 ${DOCKER_DIR}
	docker build --tag arm64/${APP_NAME} -f ${DOCKER_DIR}Dockerfile.arm64

docker-image: docker-image-linux-amd64

