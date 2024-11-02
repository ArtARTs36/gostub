try:
	gostub "./contracts/service.go" \
		--filename="services.go" \
		--per-method \
		--per-method-filename="{{ .Method.Name.Snake.Value }}.go" \
		--out="./implementations" \
		--package=implementations \
		--method-body=panic
