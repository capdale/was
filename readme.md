# Birdex was server
Birdex was server

### Config
Ref [example.yaml](./example.yaml), rename to config.yaml  


### Docker build
```powershell
docker build --progress=plain --tag was .
```

### Docker run
```powershell
docker run -d --name was -p 80:8080 was
```

### Docker Compose build
```powershell
docker compose build
```

### Docker Compose up
```powershell
docker compose up -d
```

## Protocol Buffer
[Protocol Buffer](https://protobuf.dev/overview/)  

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
/config - config type and parser  
/database - database method  
/model - database schema  
/server - initial server settings  
/static - store static file  
