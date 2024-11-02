.PHONY: build
build:
	go build ${BUILD_FLAGS} -o gostub cmd/main.go

.PHONY: try
try:
	go install ./
	cd ./examples/rpc-service && make try