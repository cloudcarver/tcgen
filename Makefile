build:
	go build -o bin/$(shell basename $(PWD)) cmd/tcgen/main.go

e2e: build
	@cd test && python3 test.py
