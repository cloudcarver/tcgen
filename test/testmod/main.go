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

func (e *MockExecutor) Set(params *fn.SetParameters) (string, error) {
	return "success", nil
}

func (e *MockExecutor) SetList(params *fn.SetListParameters) (string, error) {
	return "success", nil
}

func test(t ...bool) (string, error) {
	for _, v := range t {
		if !v {
			return "", errors.New("test failed")
		}
	}
	return "success", nil
}

func (e *MockExecutor) PlotMatrix(params *fn.PlotMatrixParameters) (string, error) {
	if params == nil {
		return "", errors.New("params is nil")
	}

	if len(params.Data) != 2 {
		return "", fmt.Errorf("data length is not correct, expected: 2, got: %d", len(params.Data))
	}

	for _, item := range params.Data {
		if len(item.Datasets) != 2 {
			return "", fmt.Errorf("datasets length is not correct, expected: 2, got: %d", len(item.Datasets))
		}
	}
	test(
		params.Data[0].Datasets[0][0].X == 1,
		params.Data[0].Datasets[0][0].Y == 1,

		params.Data[0].Datasets[0][1].X == 1,
		params.Data[0].Datasets[0][1].Y == 1,

		params.Data[0].Datasets[1][0].X == 1,
		params.Data[0].Datasets[1][0].Y == 1,

		params.Data[0].Datasets[1][1].X == 2,
		params.Data[0].Datasets[1][1].Y == 2,

		params.Data[1].Datasets[0][0].X == 2,
		params.Data[1].Datasets[0][0].Y == 2,

		params.Data[1].Datasets[0][1].X == 1,
		params.Data[1].Datasets[0][1].Y == 1,

		params.Data[1].Datasets[1][0].X == 2,
		params.Data[1].Datasets[1][0].Y == 2,

		params.Data[1].Datasets[1][1].X == 2,
		params.Data[1].Datasets[1][1].Y == 2,
	)

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

	_, err = fc.Call("PlotMatrix", `
		{"data": 
			[
				{
					"datasets": [
						[{"x": 1, "y": 1}, {"x": 1, "y": 2}],
						[{"x": 2, "y": 1}, {"x": 2, "y": 2}]
					]
				},
				{
					"datasets": [
						[{"x": 1, "y": 1}, {"x": 1, "y": 2}],
						[{"x": 2, "y": 1}, {"x": 2, "y": 2}]
					]
				}
			]
		}`)
	must(err)

	fmt.Println("✅ test passed.")
}
