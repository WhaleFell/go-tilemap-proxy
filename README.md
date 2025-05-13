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

## Reference

1. [echo web framework](https://echo.labstack.com)
