.PHONY: build
build:
	go build ${BUILD_FLAGS} -o gostub cmd/main.go

.PHONY: install
install:
	go install ./

.PHONY: try
try:
	go install ./
	cd ./examples/rpc-service && make try
	cd ./examples/rpc-service-2 && make try

.PHONY: lint
lint:
	golangci-lint run --fix
