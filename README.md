# Go Map Proxy

![Go-Map-Proxy Logo](./assets/logo.png)

Provides tile map reverse proxy and caching for private deployment of map services.

Use golang and echo web framework to implement a simple map proxy service.

**Provides the unified Google XYZ Tile map protocol.**

## Features

1. Provides the unified Google XYZ Tile map protocol across different tile map providers
2. Supports caching for improved performance and offline deployment.
3. Correcting China's tile maps (GCJ-02 coordinate system) to achieve pixel-level, lossless alignment with the universal WGS84 coordinate system.
4. Supports multiple tile map providers, including Google, OpenStreetMap, Mapbox, and more.
5. Supports custom proxy server(socks5, http, etc.) for external tile map providers.

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

# update all dependencies
go get -u ./...

# check for outdated dependencies
go list -u -m all

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
