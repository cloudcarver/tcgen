build:
	go build -o bin/$(shell basename $(PWD)) cmd/main.go

e2e:
	@cd test && python3 test.py
