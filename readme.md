# Modoo collection was server

![was_image](./docs/was.png)

<div align="center">
Modoo collection was server
</div>

# Setup

### Config

Ref [example.yaml](./example.yaml), rename to config.yaml

#### email

One of following options

- mock
  |Name|value|property|
  |---|---|---|

  This option for testing other functions

- ses
  |Name|value|property|
  |---|---|---|
  |region|us-east-2|aws ses region|
  |domain|your_domain.com|aws route53 domain|
  |id (Optional)|secret_id|aws s3 access id|
  |key (Optional)|secret_key|aws s3 access key|

  id and key pair is optional, if there is no id and key value, then server will retrieve aws ec2 temporary credential

#### storage

One of following options

- local
  |Name|value|property|
  |---|---|---|
  |baseDirectory|/root/server/tmpstorage|directory where server static files will be stored like, image etc.|

- s3
  |Name|value|property|
  |---|---|---|
  |region|us-east-2|aws s3 region|
  |bucketName|bucket0|aws s3 bucket name|
  |id (Optional)|secret_id|aws s3 access id|
  |key (Optional)|secret_key|aws s3 access key|

  id and key pair is optional, if there is no id and key value, then server will retrieve aws ec2 temporary credential

### How to run

Ref [example.yaml](./example.yaml), rename to config.yaml  
If you want to use docker network, [see this](https://docs.docker.com/network/)

```console
make docker-build
make docker-image
docker run -d -p 443:443 -v "${pwd}/secret:/server/secret" --network backnet was # set your port, image, bind mount
```

### Build (Native)

Require go

```console
# build default (linux-amd64)
make build # or make build-linux-amd64

# build all
make build-all
```

#### Options

- build-linux-amd64 (build-default)
- build-linux-arm64
- build-windows-amd64
- build-windows-arm64
- build-all

### Build (Docker)

```console
# build default (linux-amd64)
make docker-build # or make docker-build-linux-amd64

# build all
make docker-build-all
```

#### Options

- docker-build-linux-amd64 (docker-build-default)
- docker-build-linux-arm64
- docker-build-windows-amd64
- docker-build-windows-arm64
- docker-build-all

### Build Docker Image

With current ./config.yaml

```console
# build docker default image (linux-amd64)
make docker-image # or make docker-image-linux-amd64
# tag was

#build linux-arm64 docker image
make docker-image-linux-arm64
# tag arm64/was
```

#### security

Database access management  
Do not use root account to access database, make new role with restricted access

# Test and Develop

> [!WARNING]
> Do not use test setting in production!

#### Background service (mysql and redis)

[./compose.yaml](./compose.yaml)

```console
docker compose up -d
```

#### Config
