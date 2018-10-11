VERSION ?= latest

.PHONY: build
build:
	go build -o build/query-store-quay-logs main.go

.PHONY: docker-build
docker-build:
	GOOS=linux GOARCH=amd64 go build -o build/query-store-quay-logs *.go
	docker build -t quay.io/helmpack/query-store-quay-logs:$(VERSION) .

.PHONY: docker-push
docker-push:
	docker push quay.io/helmpack/query-store-quay-logs