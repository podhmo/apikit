export DEBUG ?= 0
export VERBOSE ?= 1

default: 00 01 10 11

00:
	go run $(shell echo $@*/)gen.go
01:
	go run $(shell echo $@*/)gen.go

# web
10:
	go run $(shell echo $@*/)main.go
11:
	go run $(shell echo $@*/)gen.go

check:
	go install -v ./$(shell echo 00*/)/runner/...
	go install -v ./$(shell echo 01*/)/runner/...
	go install -v ./$(shell echo 11*/)/handler/...

clean:
	rm -rf 00*/runner/
	rm -rf 01*/runner/ 01*/component
	rm -rf 11*/handler/
