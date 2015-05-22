.PHONY: cmd build

cmd:
	godep go build -o build/stat ./cmd/stat

build: Dockerfile
	docker build --no-cache -t remind101/stat .
