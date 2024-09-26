package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/cloudcarver/tcgen/internal/utils"
	"github.com/cloudcarver/tcgen/pkg/config"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

var globalTypeNameCounter = map[string]int{}

func ResetGlobalTypeNameCounter() {
	globalTypeNameCounter = map[string]int{}
}

func process(data map[string]any, onFunc func(f Function) error, onParam func(name string, params map[string]any) error) error {
	for k := range data {
		if k != "functions" {
			log.Default().Printf("[WARN] tool type %s is not supported. Skipped.", k)
		}
	}

	fns, ok := data["functions"].([]any)
	if !ok {
		return errors.New("functions is not an array")
	}

	for _, fn := range fns {
		fnData, ok := fn.(map[string]any)
		if !ok {
			return errors.New("function cannot be parsed to a map")
		}

		fnName, ok := fnData["name"].(string)
		if !ok {
			return errors.New("function name cannot be parsed to a string")
		}

		var description string

		if _, ok := fnData["description"]; ok {
			description, ok = fnData["description"].(string)
			if !ok {
				return errors.New("function description cannot be parsed to a string")
			}
		}

		// parse parameters
		if _, ok := fnData["parameters"]; !ok {
			return errors.New("parameters is missing when function type is object")
		}
		parameters, ok := fnData["parameters"].(map[string]any)
		if !ok {
			return errors.New("parameters cannot be parsed to a map")
		}

		structName := addGlobalType(fmt.Sprintf("%sParameters", utils.UpperFirst(fnName)))
		if err := onParam(structName, parameters); err != nil {
			return err
		}

		if err := onFunc(Function{
			Name:          fnName,
			Description:   description,
			ParameterType: structName,
		}); err != nil {
			return err
		}
	}
	return nil
}

func fnNameToAPIPath(fnName string) string {
	rtn := ""
	rtn += strings.ToLower(string(fnName[0]))
	for _, c := range fnName[1:] {
		if c >= 'A' && c <= 'Z' {
			rtn += "_" + strings.ToLower(string(c))
		} else {
			rtn += string(c)
		}
	}
	return rtn
}

func descriptionToComment(description string) string {
	description = strings.Trim(description, " \n\t\r")
	var rtn = ""
	var arr = strings.Split(description, "\n")
	for i, line := range arr {
		rtn += "// " + line
		if i != len(arr)-1 {
			rtn += "\n"
		}
	}
	return indent(rtn, 4)
}

func GenerateOpenAPISpec(original []byte, data map[string]any, cfg *config.OpenAPI) (string, error) {

	type Parameters struct {
		Name string
		Spec string
	}

	pathPrefix := "/tcgen"
	if cfg.Paths.Prefix != "" {
		pathPrefix = cfg.Paths.Prefix
	}

	if original == nil {
		original = []byte("openapi: 3.1.0\ninfo:\n  version: 1.0.0\n  title:  Tool Call Server API\npaths:")
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(original)
	if err != nil {
		return "", err
	}

	functions := []Function{}
	parameters := []Parameters{}

	onFunc := func(f Function) error {
		if cfg.Paths.Skip {
			return nil
		}

		functions = append(functions, f)
		if doc.Paths == nil {
			doc.Paths = openapi3.NewPaths()
		}
		var descp = "OK"
		res := &openapi3.Responses{}
		res.Set("200", &openapi3.ResponseRef{
			Value: &openapi3.Response{
				Description: &descp,
			},
		})
		doc.Paths.Set(fmt.Sprintf("%s/%s", pathPrefix, fnNameToAPIPath(f.Name)), &openapi3.PathItem{
			Post: &openapi3.Operation{
				Description: f.Description,
				RequestBody: &openapi3.RequestBodyRef{
					Value: &openapi3.RequestBody{
						Content: map[string]*openapi3.MediaType{
							"application/json": {
								Schema: &openapi3.SchemaRef{
									Ref: fmt.Sprintf("#/components/schemas/%s", f.ParameterType),
								},
							},
						},
					},
				},
				Responses: res,
			},
		})
		return nil
	}

	genComponentsTemplate := `components:
  schemas:
{{range .Parameters}}
    {{.Name}}:
{{.Spec}}
{{end}}`

	type ComponentsTemplateVars struct {
		Parameters []Parameters
	}

	onParam := func(name string, params map[string]any) error {
		buf := bytes.Buffer{}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		err := enc.Encode(params)
		if err != nil {
			return err
		}

		parameters = append(parameters, Parameters{
			Name: name,
			Spec: indent(buf.String(), 6),
		})
		return nil
	}

	componentsTemplate, err := template.New("components").Parse(genComponentsTemplate)
	if err != nil {
		return "", err
	}

	if err := process(data, onFunc, onParam); err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := componentsTemplate.Execute(buf, ComponentsTemplateVars{
		Parameters: parameters,
	}); err != nil {
		return "", err
	}

	componentsDoc, err := loader.LoadFromData(buf.Bytes())
	if err != nil {
		return "", err
	}
	if doc.Components == nil {
		doc.Components = componentsDoc.Components
	} else {
		for k, v := range componentsDoc.Components.Schemas {
			doc.Components.Schemas[k] = v
		}
	}

	out, err := marshalOpenAPIDoc(doc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func marshalWithIndent(data any, indent int) ([]byte, error) {
	buf := bytes.Buffer{}
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func marshalOpenAPIDoc(doc *openapi3.T) ([]byte, error) {
	rtn := ""
	raw, err := doc.MarshalYAML()
	if err != nil {
		return nil, err
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("failed to marshal openapi doc")
	}
	openapi, err := marshalWithIndent(data["openapi"], 2)
	if err != nil {
		return nil, err
	}
	info, err := marshalWithIndent(data["info"], 2)
	if err != nil {
		return nil, err
	}
	paths, err := marshalWithIndent(data["paths"], 2)
	if err != nil {
		return nil, err
	}

	components, err := marshalWithIndent(data["components"], 2)
	if err != nil {
		return nil, err
	}
	rtn += "openapi: \"" + strings.Trim(string(openapi), "\n\t\r ") + "\"\n\n" +
		"info:\n" + indent(string(info), 2) + "\n" +
		"paths:\n" + indent(string(paths), 2) + "\n" +
		"components:\n" + indent(string(components), 2) + "\n"
	return []byte(rtn), nil
}

func indent(s string, spaces int) string {
	indent := strings.Repeat(" ", spaces)
	return indent + strings.ReplaceAll(s, "\n", "\n"+indent)
}

func GenerateToolInterfaces(packageName string, data map[string]any) (string, error) {
	var structDef string
	functions := []Function{}

	onFunc := func(f Function) error {
		functions = append(functions, f)
		return nil
	}

	onParam := func(name string, params map[string]any) error {
		def, err := parseObjectToStruct(name, params)
		if err != nil {
			return err
		}
		structDef += def + "\n"
		return nil
	}

	tcTemplate, err := template.New("toolCalls").Parse(genToolCallsTemplate)
	if err != nil {
		return "", err
	}

	if err := process(data, onFunc, onParam); err != nil {
		return "", err
	}

	for i := range functions {
		functions[i].Description = descriptionToComment(functions[i].Description)
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tcTemplate.Execute(buf, CodeTemplateVars{
		PackageName: packageName,
		StructDefs:  structDef,
		Functions:   functions,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func addGlobalType(name string) string {
	if _, ok := globalTypeNameCounter[name]; ok {
		globalTypeNameCounter[name]++
		return fmt.Sprintf("%s%d", name, globalTypeNameCounter[name])
	} else {
		globalTypeNameCounter[name] = 0
		return name
	}
}

func parseArrayToStruct(name string, data map[string]any) (string, string, error) {
	items, ok := data["items"].(map[string]any)
	if !ok {
		return "", "", errors.New("items cannot be parsed to a map")
	}
	itemsType, ok := items["type"].(string)
	if !ok {
		return "", "", errors.New("items type cannot be parsed to a string")
	}

	if itemsType == "object" {
		if _, ok := items["properties"]; !ok {
			return "[]any", "", nil
		}
		propStructName := utils.UpperFirst(name) + "Item"
		propStructDef, err := parseObjectToStruct(propStructName, items)
		if err != nil {
			return "", "", err
		}
		return "[]" + propStructName, propStructDef, nil
	} else if itemsType == "array" {
		propStructName, propStructDef, err := parseArrayToStruct(name, items)
		if err != nil {
			return "", "", err
		}
		return "[]" + propStructName, propStructDef, nil
	} else {
		return "[]" + typeMap(itemsType), "", nil
	}
}

func typeMap(typeName string) string {
	switch typeName {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return typeName
	}
}

// return struct name, struct definition, error
func parseObjectToStruct(structName string, object map[string]any) (string, error) {
	var ok bool
	var requiredFields = map[string]struct{}{}
	var properties map[string]any
	var structDef string

	if _, ok := object["properties"]; !ok {
		return "", nil
	}

	properties, ok = object["properties"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("properties %v cannot be parsed to map[string]map[string]any", object["properties"])
	}

	if _, ok := object["required"]; ok {
		required, ok := object["required"].([]any)
		if !ok {
			return "", fmt.Errorf("required %v cannot be parsed to a string array", object["required"])
		}
		for _, r := range required {
			if _, ok := properties[r.(string)]; !ok {
				return "", fmt.Errorf("required field %s is not in properties", r)
			}
			requiredFields[r.(string)] = struct{}{}
		}
	}

	tmpl, err := template.New("struct").Parse(structTemplate)
	if err != nil {
		return "", err
	}

	fields := []Field{}

	for propName, propRaw := range properties {
		prop, ok := propRaw.(map[string]any)
		if !ok {
			return "", fmt.Errorf("property %s cannot be parsed to a map", propName)
		}
		propType, ok := prop["type"].(string)
		if !ok {
			return "", errors.New("property type cannot be parsed to a string")
		}

		var propDescription string
		if _, ok := prop["description"]; ok {
			propDescription, ok = prop["description"].(string)
			if !ok {
				return "", errors.New("property description cannot be parsed to a string")
			}
		}

		_, isRequired := requiredFields[propName]

		if propType == "object" {
			if _, ok := prop["properties"]; !ok {
				propType = "any"
			} else {
				propStructName := addGlobalType(utils.UpperFirst(propName))
				propStructDef, err := parseObjectToStruct(propStructName, prop)
				if err != nil {
					return "", err
				}
				propType = propStructName
				structDef += propStructDef + "\n"
			}
		} else if propType == "array" {
			propStructName, propStructDef, err := parseArrayToStruct(propName, prop)
			if err != nil {
				return "", err
			}
			propType = propStructName
			structDef += propStructDef + "\n"
		} else {
			propType = typeMap(propType)
		}

		fields = append(fields, Field{
			Name:        utils.UpperFirst(propName),
			Type:        utils.IfElse(isRequired || strings.HasPrefix(propType, "[]"), "", "*") + propType,
			Description: descriptionToComment(propDescription),
			Tag:         "`json:\"" + propName + "\" yaml:\"" + propName + "\"`",
		})
	}

	templateVars := StructTemplateVars{
		StructName: structName,
		Fields:     fields,
	}
	buf := bytes.NewBuffer([]byte{})

	if err := tmpl.Execute(buf, templateVars); err != nil {
		return "", err
	}

	return structDef + "\n" + buf.String(), nil
}
