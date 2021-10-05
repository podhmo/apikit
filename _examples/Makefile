export DEBUG = 1
default: 00 01 10 11 12

00:
	go run $(shell echo $@*/)main.go
01:
	go run $(shell echo $@*/)main.go

# web
10:
	go run $(shell echo $@*/)main.go
11:
	go run $(shell echo $@*/)main.go # todo: remove?
12:
	go run $(shell echo $@*/)main.go

check:
	go install -v ./$(shell echo 00*/)/runner/...
	go install -v ./$(shell echo 01*/)/runner/...
	go install -v ./$(shell echo 11*/)/handler/...
	go install -v ./$(shell echo 12*/)/handler/...

clean:
	rm -rf 00*/runner/
	rm -rf 01*/runner/ 01*/component
	rm -rf 11*/handler/
	rm -rf 12*/handler/	