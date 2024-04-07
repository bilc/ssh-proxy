.PHONY: static-proxy  

example: example/*/*go
	@go build -o example-client example/client/*.go
	@go build -o example-server example/server/*.go
	@go build -o example-proxy example/proxy/*.go

proxy-k8s:
    go build -o proxy-k8s cmd/proxy-k8s/*go
static-proxy: cmd/static-proxy/*go
	@go build -o static-proxy cmd/static-proxy/*go

clean:
	rm example-client example-server example-proxy