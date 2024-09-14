package core

type Field struct {
	Description string
	Name        string
	Type        string
	Tag         string
}

type StructTemplateVars struct {
	StructName string
	Fields     []Field
}

type Function struct {
	Name          string
	Description   string
	ParameterType string
}

type CodeTemplateVars struct {
	FunctionsDef string
	PackageName  string
	StructDefs   string
	Functions    []Function
}
