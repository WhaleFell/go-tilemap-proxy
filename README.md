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

## Map Providers Supported

1. Google Satellite Map (谷歌卫星)
2. Google Hybrid Map (Satellite + Road)
3. Google Hybrid Offset Map (谷歌混合 GCJ02 加偏)
4. Bing Satellite Map (必应卫星)
5. OpenStreetMap (开源街道地图)
6. OpenStreetMap Cycle Map (等高线)
7. Trace Strack Topo Map
8. Amap Road Map (高德路网[纠偏后])
9. TianDiTu Satellite Map (天地图卫星)
10. TianDiTu Road Map (天地图路网[CGCS2000])
11. MapHere Satellite Map (MapHere 卫星)
12. Map Tiler Contour (等高线)
13. (ESRI) Arcgis Satelite Map (ESRI 卫星)
14. Tencent Satellite Map (腾讯卫星[GCJ02])
15. Huawei Road Map (Petal Map 华为花瓣地图[GCJ02])

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

## Deployment

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

Best practice: Nginx (Reverse Proxy) HTTPS + docker-compose + CDN

```shell
# copy docker-compose.yaml and nginx.conf
# self-signed certificate for testing
mkdir ssl

openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout ./ssl/privkey.pem \
    -out ./ssl/fullchain.pem \
    -subj "/C=CN/ST=State/L=City/O=Organization/CN=localhost"


```

## Usage

GET: `/map/list/` - List all available map sources

Response:

```json
{
  "code": 200,
  "message": "Get tile map source list success",
  "data": [
    {
      "name": "Google Pure Satellite",
      "id": "google_pure_satellite",
      "min_zoom": 0,
      "max_zoom": 20,
      "map_type": "raster",
      "map_size": 256,
      "coordinate_type": "EPSG:4326",
      "content_type": "image/png"
    },
    ...
  ]
}
```

GET: `/map/{map_id}/{x}/{y}/{z}/?cache=true` - Get tile map by map type and XYZ coordinates, cache is enabled by default.
Example: `/map/google_pure_satellite/1200/1343/11/`

GET: `/map/testpage/` - A simple test page for the map service.

---

GET: `/health` - Health check endpoint.
GET: `/proxy/?url=<url>` - Proxy specified URL. example: `/proxy/?url=https://www.google.com`

GET: `/systemInfo` - Get system information.
Response:

```json
{
  "code": 200,
  "message": "Get system info success",
  "data": {
    "os": "windows",
    "architecture": "amd64",
    "go_version": "go1.24.6",
    "executorPath": "C:\\Users\\username\\AppData\\Local\\Temp\\go-build2630162718\\b001\\exe\\server.exe",
    "executableDir": "C:\\Users\\username\\AppData\\Local\\Temp\\go-build2630162718\\b001\\exe",
    "workingDir": "E:\\TEMP\\Golang-Project\\go-map-proxy",
    "cpu": {
      "type": "AMD Ryzen 5 5500U with Radeon Graphics",
      "cores": 12,
      "thread": 6,
      "percent": 4.43
    },
    "memory": { "total": 15.34, "used": 13.71, "used_percent": 89.37 }
  }
}
```

## Reference

1. [echo web framework](https://echo.labstack.com)
