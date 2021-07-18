build:
	GO111MODULE=on GOOS=linux GOARCH=amd64 go build -mod vendor -o ping-pong -v main.go

build2:
	CGO_ENABLED=0 GOOS=linux go build -tags=jsoniter -a -installsuffix cgo -o ping-pong .

test: 
	go test -coverprofile=coverage/cover.out ./...
	go tool cover -func=coverage/cover.out | grep total
	go tool cover -html=coverage/cover.out -o coverage/cover.html

tests:
	go test ./...

lint:
	golangci-lint run --skip-dirs='(mocks|vendor|1)' --fix -E gofmt,gofumpt,goimports

fmt:
	go fmt ./...

vendor:
	go mod vendor

mock:
	rm -rf ./vendor
	mockery --all --keeptree
	export GOSUMDB=off && go mod vendor

remock:
	rm -rf ./mocks && rm -rf ./vendor
	mockery --all --keeptree
	export GOSUMDB=off && go mod vendor
