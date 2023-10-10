# Birdex was server
Birdex was server

### Config
Ref [example.yaml](./example.yaml), rename to config.yaml  


### How to run
Ref [example.yaml](./example.yaml), rename to config.yaml  
If you want to use docker network, [see this](https://docs.docker.com/network/)  
```powershell
make docker-build
make docker-image
docker run -d -p 8080:8080 -v "${pwd}/secret:/server/secret" --network backnet was # set your port, image, bind mount
```


### Build (Native)
Require go  
```powershell
# build default (linux-amd64)
make build # or make build-linux-amd64

# build linux-arm64
make build-linux-arm64

# build windows-amd64
make build-widnows-amd64

# build windows-arm64
make build-windows-arm64

# build all
make build-all
```

### Build (Docker)
```powershell
# build default (linux-amd64)
make docker-build # or make docker-build-linux-amd64

# build linux-arm64
make docker-build-linux-arm64

# build windows-amd64
make docker-build-widnows-amd64

# build windows-arm64
make docker-build-windows-arm64

# build all
make docker-build-all
```

### Build Docker Image
With current ./config.yaml  
```powershell
# build docker default image (linux-amd64)
make docker-image # or make docker-image-linux-amd64
# tag was

#build linux-arm64 docker image 
make docker-image-linux-arm64
# tag arm64/was
```

### Background service for development
[./compose.yaml](./compose.yaml)  
```powershell
docker compose up -d
```

### File structure
```
was
 ┣ api
 ┃ ┣ collect
 ┃ ┃ ┗ collect.go
 ┃ ┗ api.go
 ┣ config
 ┃ ┗ config.go
 ┣ database
 ┃ ┗ new.go
 ┣ model
 ┣ server
 ┃ ┗ server.go
 ┣ static
 ┣ .dockerignore
 ┣ .gitignore
 ┣ compose.yaml
 ┣ Dockerfile
 ┣ example.yaml
 ┣ go.mod
 ┣ go.sum
 ┣ main.go
 ┗ readme.md
```
/api - collection of api  
/config - config parser  
/database - database functions    
/model - database schema  
/server - initial server settings  
/static - store static file  

## Proto
[Check proto here](https://github.com/capdale/rpc-protocol)  
