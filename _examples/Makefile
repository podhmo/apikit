export DEBUG ?= 0
export VERBOSE ?= 1

default: 00 01 02 10 11 12

00:
	go run $(shell echo $@*/)gen.go
01:
	go run $(shell echo $@*/)gen.go
02:
	go run $(shell echo $@*/)gen.go

# web
10:
	go run $(shell echo $@*/)main.go
11:
	go run $(shell echo $@*/)gen.go
12:
	go run $(shell echo $@*/)gen.go

check:
	go install -v ./$(shell echo 00*/)/runner/...
	go install -v ./$(shell echo 01*/)/runner/...
	go install -v ./$(shell echo 02*/)/component/...
	go install -v ./$(shell echo 11*/)/handler/...
	go install -v ./$(shell echo 12*/)/handler/...

# apikit
init-check:
	rm -rf foo
	@echo ----------------------------------------
	VERBOSE=1 go run ../cmd/apikit init m/foo
	@echo ----------------------------------------
	VERBOSE=1 go run foo/gen.go
	go install -v ./foo/...	
	# rm -r foo
.PHONY: init-check

clean:
	rm -rf 00*/runner/
	rm -rf 01*/runner/ 01*/component
	rm -rf 02*/component
	rm -rf 11*/handler/
	rm -rf 12*/handler/