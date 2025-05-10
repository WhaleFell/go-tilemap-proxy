# Go Map Proxy

Provides tile map reverse proxy and caching for private deployment of map services.

Use golang and echo web framework to implement a simple map proxy service.

## Development

```bash
# enable go module
export GO111MODULE=on

go mod init go-map-proxy

# install dependencies
go get github.com/labstack/echo/v4
go get github.com/spf13/viper
go mod tiny

# run
go run ./cmd/server

# build
go build -o go-map-proxy ./cmd/server


```

## Reference

1. [echo web framework](https://echo.labstack.com)
