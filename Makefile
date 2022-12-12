BINARY_NAME="api_gateway"
WINDOWS_NAME="server.exe"

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o build/${BINARY_NAME} ./cmd/gateway

windows:
	GOOS=windows GOARCH=amd64 go build -a -o build/${WINDOWS_NAME} ./cmd/gateway

mac:
	GOOS=darwin GOARCH=amd64 go build -o build/${BINARY_NAME} ./cmd/gateway

clean:
	rm -rf build/*

test:
	go test ./...
