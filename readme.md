# Birdex was server
Birdex was server

### Config
Ref [example.yaml](./example.yaml), rename to config.yaml  


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
- docker network -net backnet
- port
    - redis: 6379:6379
    - mysql: 3306:3306
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
