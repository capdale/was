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