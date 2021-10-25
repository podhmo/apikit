default: test lint

test:
	go test ./...
	cd web/webgen/gen-chi/webruntime && go test ./...
.PHONY: test

lint:
	go vet ./...
.PHONY: lint

install:
	go install -v ./cmd/apikit
.PHONY: install	