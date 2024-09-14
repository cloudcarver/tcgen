package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/cloudcarver/tcgen/internal/utils"
)

func Generate(data map[string]any) (string, error) {
	for k := range data {
		if k != "functions" {
			log.Default().Printf("[WARN] tool type %s is not supported.", k)
		}
	}

	fnTemplate, err := template.New("functions").Parse(genTemplate)
	if err != nil {
		return "", err
	}

	fns, ok := data["functions"].([]any)
	if !ok {
		return "", errors.New("functions is not an array")
	}

	functions := []Function{}
	var structDef string
	for _, fn := range fns {
		fnData, ok := fn.(map[string]any)
		if !ok {
			return "", errors.New("function cannot be parsed to a map")
		}

		fnName, ok := fnData["name"].(string)
		if !ok {
			return "", errors.New("function name cannot be parsed to a string")
		}

		var description string

		if _, ok := fnData["description"]; ok {
			description, ok = fnData["description"].(string)
			if !ok {
				return "", errors.New("function description cannot be parsed to a string")
			}
		}

		// parse parameters
		if _, ok := fnData["parameters"]; !ok {
			return "", errors.New("parameters is missing when function type is object")
		}
		parameters, ok := fnData["parameters"].(map[string]any)
		if !ok {
			return "", errors.New("parameters cannot be parsed to a map")
		}
		paramType, paramsDef, err := parseObjectToStruct(fmt.Sprintf("%sParameters", utils.UpperFirst(fnName)), parameters)
		if err != nil {
			return "", err
		}

		structDef += paramsDef + "\n"

		functions = append(functions, Function{
			Name:          fnName,
			Description:   description,
			ParameterType: paramType,
		})
	}

	functionsDef, err := json.Marshal(fns)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := fnTemplate.Execute(buf, CodeTemplateVars{
		FunctionsDef: string(functionsDef),
		PackageName:  "fn",
		StructDefs:   structDef,
		Functions:    functions,
	}); err != nil {
		return "", err
	}

	return buf.String(), nil
}

var globalTypeNameCounter = map[string]int{}

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
		propStructName, propStructDef, err := parseObjectToStruct(utils.UpperFirst(name)+"Item", items)
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
func parseObjectToStruct(name string, object map[string]any) (string, string, error) {
	var ok bool
	var requiredFields = map[string]struct{}{}
	var properties map[string]any
	var structName string
	var structDef string

	structName = addGlobalType(name)

	if _, ok := object["properties"]; !ok {
		return "", "", errors.New("properties is missing")
	}

	properties, ok = object["properties"].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("properties %v cannot be parsed to map[string]map[string]any", object["properties"])
	}

	if _, ok := object["required"]; ok {
		required, ok := object["required"].([]any)
		if !ok {
			return "", "", fmt.Errorf("required %v cannot be parsed to a string array", object["required"])
		}
		for _, r := range required {
			if _, ok := properties[r.(string)]; !ok {
				return "", "", fmt.Errorf("required field %s is not in properties", r)
			}
			requiredFields[r.(string)] = struct{}{}
		}
	}

	tmpl, err := template.New("struct").Parse(structTemplate)
	if err != nil {
		return "", "", err
	}

	fields := []Field{}

	for propName, propRaw := range properties {
		prop, ok := propRaw.(map[string]any)
		if !ok {
			return "", "", fmt.Errorf("property %s cannot be parsed to a map", propName)
		}
		propType, ok := prop["type"].(string)
		if !ok {
			return "", "", errors.New("property type cannot be parsed to a string")
		}

		var propDescription string
		if _, ok := prop["description"]; ok {
			propDescription, ok = prop["description"].(string)
			if !ok {
				return "", "", errors.New("property description cannot be parsed to a string")
			}
		}

		_, isRequired := requiredFields[propName]

		if propType == "object" {
			propStructName, propStructDef, err := parseObjectToStruct(propName, prop)
			if err != nil {
				return "", "", err
			}
			propType = propStructName
			structDef += propStructDef + "\n"
		} else if propType == "array" {
			propStructName, propStructDef, err := parseArrayToStruct(propName, prop)
			if err != nil {
				return "", "", err
			}
			propType = propStructName
			structDef += propStructDef + "\n"
		} else {
			propType = typeMap(propType)
		}

		fields = append(fields, Field{
			Name:        utils.UpperFirst(propName),
			Type:        utils.IfElse(isRequired || strings.HasPrefix(propType, "[]"), "", "*") + propType,
			Description: propDescription,
			Tag:         "`json:\"" + propName + "\" yaml:\"" + propName + "\"`",
		})
	}

	templateVars := StructTemplateVars{
		StructName: structName,
		Fields:     fields,
	}
	buf := bytes.NewBuffer([]byte{})

	if err := tmpl.Execute(buf, templateVars); err != nil {
		return "", "", err
	}

	return structName, structDef + "\n" + buf.String(), nil
}
