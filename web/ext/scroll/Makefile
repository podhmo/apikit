default: clean gen check
	
gen:
	VERBOSE=1 go run ./_examples/gen.go

check:
	go install -v ./_examples/runtime

clean:
	rm -rf ./_examples/runtime