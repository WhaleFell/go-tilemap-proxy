# Go Map Proxy

Provides tile map reverse proxy and caching for private deployment of map services.

Use golang and echo web framework to implement a simple map proxy service.

**Provides the unified Google XYZ Tile map protocol.**

## Development

```bash
# enable go module
export GO111MODULE=on

go mod init go-map-proxy

# check golang env
go env

# install dependencies
go get github.com/labstack/echo/v4
go get github.com/spf13/viper
go mod tiny

# run
go run ./cmd/server

# build
go build -o go-map-proxy ./cmd/server


# build with optimization
go build -o ./bin/server.exe -trimpath -buildvcs=false -ldflags="-s -w -buildid= -checklinkname=0" -v ./cmd/server

```

## deployment

Recommend using docker to deploy the service.

```bash
# build docker image
docker build \
  --build-arg HTTP_PROXY=http://172.23.0.1:10808 \
  --build-arg HTTPS_PROXY=http://172.23.0.1:10808 \
  -f ./Dockerfile \
  -t whalefell/map-server:test .

# test run
docker run -it --rm whalefell/map-server:test bash

# set tag to latest
docker tag whalefell/map-server:test whalefell/map-server:latest

# push to docker hub
docker push whalefell/map-server:latest
docker push whalefell/map-server:test

# run docker container
docker run -d \
  --name map-server \
  -p 8080:8080 \
  -v /path/to/config:/app/config \
  -v /path/to/cache:/cache \
  whalefell/map-server:latest

# run docker by docker-compose
docker compose up -d

# build docker-compose image
docker compose build
```

## Reference

1. [echo web framework](https://echo.labstack.com)
