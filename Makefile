test:
	go test ./...
.PHONY: test

lint:
	go vet ./...
.PHONY: lint

install:
	go install -v ./cmd/apikit
.PHONY: install	