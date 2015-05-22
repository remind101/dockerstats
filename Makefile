.PHONY: cmd build

cmd:
	godep go build -o build/stats ./cmd/stats

build: Dockerfile
	docker build --no-cache -t remind101/dockerstats .
