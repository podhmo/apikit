export DEBUG = 1
default: 00 01 10 11

00:
	go run $(shell echo $@*/)main.go
01:
	go run $(shell echo $@*/)main.go

# web
10:
	go run $(shell echo $@*/)main.go
11:
	go run $(shell echo $@*/)main.go

check:
	go install -v ./$(shell echo 00*/)/...
	go install -v ./$(shell echo 01*/)/...

clean:
	rm -rf 00*/runner/
	rm -rf 01*/runner/ 01*/component
	rm -rf 11*/handler/