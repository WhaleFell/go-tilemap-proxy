# set env 
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# build command
go build -o ./bin/server -trimpath -buildvcs=false -ldflags="-s -w -buildid= -checklinkname=0" -v ./cmd/server