BIN=$(shell basename $(PWD))

build:
	go build -o bin/$(BIN) cmd/tcgen/main.go

install: build
	@cp bin/$(BIN) $(GOPATH)/bin

e2e: build
	@cd test && python3 test.py
