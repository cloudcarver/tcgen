package core

var structTemplate = `type {{.StructName}} struct {
	{{range .Fields}}
{{.Description}}
	{{.Name}} {{.Type}} {{.Tag}}
	{{end}}
}`

var genToolCallsTemplate = `// This file is generated by tools, DO NOT EDIT.
package {{.PackageName}}

import (
	"encoding/json"
	"fmt"
)
{{.StructDefs}}{{range .Functions}}
func (r *{{.ParameterType}}) Parse(raw string) error {
	return json.Unmarshal([]byte(raw), r)
}
{{end}}
type FunctionExecutorInterface interface {
	{{range .Functions}}
{{.Description}}
	{{.Name}}(params *{{.ParameterType}}) (string, error)
	{{end}}
}
type FunctionCaller struct {
	executor FunctionExecutorInterface
}

func NewFunctionCaller(executor FunctionExecutorInterface) *FunctionCaller {
	return &FunctionCaller{
		executor: executor,
	}
}

func (f *FunctionCaller) Call(fnName string, paramJSON string) (string, error) {
	switch fnName {
	{{range .Functions}}
	case "{{.Name}}":
		var params {{.ParameterType}}
		if err := params.Parse(paramJSON); err != nil {
			return "", fmt.Errorf("failed to parse {{.Name}} parameters: %w", err)
		}
		return f.executor.{{.Name}}(&params)
		{{end}}
	default:
		return "", fmt.Errorf("unknown function %s", fnName)
	}
}
`
