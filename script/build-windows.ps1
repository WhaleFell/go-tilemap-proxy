# set env 
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# build command
go build -o ./bin/server.exe -trimpath -buildvcs=false -ldflags="-s -w -buildid= -checklinkname=0" -v ./cmd/server