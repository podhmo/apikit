export DEBUG = 1
default: 00 01

00:
	go run $(shell echo $@*/)main.go
01:
	go run $(shell echo $@*/)main.go

x00:
	go run $(shell echo $@*/)main.go

check:
	go install -v ./$(shell echo 00*/)/...
	go install -v ./$(shell echo 01*/)/...