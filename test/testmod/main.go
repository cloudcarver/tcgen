package main

import (
	"errors"
	"fmt"
	"testmod/fn"
)

type MockExecutor struct {
}

func (e *MockExecutor) RunPython(params *fn.RunPythonParameters) (string, error) {
	if params == nil {
		return "", errors.New("params is nil")
	}

	expected := "print('hello')"
	if params.Code != expected {
		return "", fmt.Errorf("code is not correct, expected: %s, got: %s", expected, params.Code)
	}

	return "success", nil
}

func (e *MockExecutor) WebSearch(params *fn.WebSearchParameters) (string, error) {
	if params == nil {
		return "", errors.New("params is nil")
	}

	expected := "what's new?"
	if params.Query != expected {
		return "", fmt.Errorf("query is not correct, expected: %s, got: %s", expected, params.Query)
	}

	return "success", nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fc := fn.NewFunctionCaller(&MockExecutor{})
	_, err := fc.Call("RunPython", `{"code": "print('hello')"}`)
	must(err)

	_, err = fc.Call("WebSearch", `{"query": "what's new?"}`)
	must(err)

	fmt.Println("✅ test passed.")
}
