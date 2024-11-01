.PHONY: build
build:
	go build ${BUILD_FLAGS} -o gostub cmd/main.go

.PHONY: try
try:
	make build && \
		./gostub "./../users.go" --per-type --per-method-filename="{{ .Interface.Name.Snake.Value }}_{{ .Method.Name.Snake.Value }}_stub.go" --out="./out" --method-body=panic
