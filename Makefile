test:
	go test ./...
.PHONY: test

lint:
	go vet ./...
.PHONY: lint
