# Tool Calls Code Generator

## Quick Start

1. Define tool calls in `test.yaml`:
    ```yaml
    functions:
    - name: RunPython
        description: Python code execution
        parameters:
        type: object
        required: [code]
        properties:
            code:
            type: string
            description: the Python code you want to run
    ```

2. Run the code generator:
    ```shell
    tcgen -path test.yaml > fn_gen.go
    ```

3. Then you can get the parameter types and the function signature in `fn_gen.go`, all you need to 
do is implementing the interfaces and then you can call tools with plain text returned by LLMs:

    ```go
    type MyExecutor struct {
    }

    func (e *MyExecutor) RunPython(params *RunPythonParameters) (string, error) {
        // implement your logic here
        return "", nil
    }

    func main() {
        fc := NewFunctionCaller(&MyExecutor{})
        _, err := fc.Call("RunPython", `{"code": "print('hello')"}`)
    }
    ```
