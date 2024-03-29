package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
)

const generatorLabel = "// apigen:api"

var serveHTTPRoutes = map[string]string{}

// handlerConfig - settings of method
type handlerConfig struct {
	URL    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}

type routeTemplateData struct {
	URL      string
	FuncName string
}

type serveHTTPTemplateData struct {
	StructName string
	Router     string
}

type wrapperTemplateData struct {
	ReceiverType string
	ParamsType   string
	FuncName     string
	MethodCheck  string
	AuthCheck    string
}

type typeNameData struct {
	TypeName string
}

type unpackData struct {
	QueryName  string
	FieldName  string
	DefaultVal string
}

// getReceiverType - get type name of receiver struct
func getReceiverType(fd *ast.FuncDecl) string {
	st, ok1 := fd.Recv.List[0].Type.(*ast.StarExpr)
	if ok1 {
		nodeType, ok2 := st.X.(*ast.Ident)
		if ok2 {
			return nodeType.Name
		}
	}
	return ""
}

// getParamType - get type name of function parameter
func getParamType(fd *ast.FuncDecl) string {
	i, ok := fd.Type.Params.List[1].Type.(*ast.Ident)
	if ok {
		return i.Name
	}
	return ""
}

// createWrapper - generate http wrapper code
func createWrapper(fd *ast.FuncDecl, config *handlerConfig) (string, error) {
	rt := getReceiverType(fd)
	pt := getParamType(fd)
	var tmp bytes.Buffer
	data := wrapperTemplateData{
		ReceiverType: rt,
		ParamsType:   pt,
		FuncName:     fd.Name.Name,
	}
	if config.Method == "POST" {
		data.MethodCheck = methodCheckTemplate
	}
	if config.Auth {
		data.AuthCheck = authCheckTemplate
	}
	err := wrapperTpl.Execute(&tmp, data)
	if err != nil {
		return "", err
	}
	return tmp.String(), nil
}

// createRoute - generate router code
func createRoute(fd *ast.FuncDecl, config *handlerConfig) error {
	typeName := getReceiverType(fd)
	var tmp bytes.Buffer
	err := routeTpl.Execute(&tmp, routeTemplateData{config.URL, fd.Name.Name})
	if err != nil {
		return err
	}
	serveHTTPRoutes[typeName] += tmp.String()
	return nil
}

// parseStructTag - parse struct tag and return options
func parseStructTag(tag string) map[string]string {
	result := map[string]string{}
	if tag == "" {
		return result
	}
	options := strings.Split(tag, ",")
	for _, v := range options {
		tmp := strings.Split(v, "=")
		if len(tmp) > 1 {
			result[tmp[0]] = tmp[1]
		} else {
			result[tmp[0]] = ""
		}
	}
	return result
}

// getQueryName - get name of query parameter from struct tag
func getQueryName(options map[string]string, fieldName string) string {
	if queryName, ok := options["paramname"]; ok {
		return fmt.Sprintf(`%s`, queryName)
	}
	return fmt.Sprintf(`%s`, strings.ToLower(fieldName))
}

// getDefaultValue - get default value from struct tag
func getDefaultValue(options map[string]string, typeName string) string {
	v, ok := options["default"]
	if !ok {
		if typeName == "string" {
			return `""`
		}
		return "0"
	}
	if typeName == "string" {
		return fmt.Sprintf(`"%s"`, v)
	}
	return fmt.Sprintf("%v", v)
}

// createUnpackCode - create unpack code for single field
func createUnpackCode(field *ast.Field, options map[string]string) (string, error) {
	fieldName := field.Names[0].Name
	fieldType := field.Type.(*ast.Ident).Name
	var tmp bytes.Buffer
	data := unpackData{
		QueryName: getQueryName(options, fieldName),
		FieldName: fieldName,
	}
	switch fieldType {
	case "int":
		data.DefaultVal = getDefaultValue(options, "int")
		err := intTpl.Execute(&tmp, data)
		if err != nil {
			return "", err
		}
		return tmp.String(), nil
	case "string":
		data.DefaultVal = getDefaultValue(options, "string")
		err := strTpl.Execute(&tmp, data)
		if err != nil {
			return "", err
		}
		return tmp.String(), nil
	default:
		return "", fmt.Errorf("unsupported, %+v", fieldType)
	}
}

// createUnpackCode - create unpack code for field
func createValidateCode(field *ast.Field, options map[string]string) string {
	var code string
	fieldName := field.Names[0].Name
	filedType := field.Type.(*ast.Ident).Name
	queryName := getQueryName(options, fieldName)

	if _, ok := options["required"]; ok {
		if filedType == "int" {
			code += fmt.Sprintf(requiredTemplateInt, fieldName, queryName)
		} else if filedType == "string" {
			code += fmt.Sprintf(requiredTemplateString, fieldName, queryName)
		}
	}

	if min, ok := options["min"]; ok {
		if filedType == "int" {
			code += fmt.Sprintf(minTemplateInt, fieldName, min, queryName, min)
		} else if filedType == "string" {
			code += fmt.Sprintf(minTemplateString, fieldName, min, queryName, min)
		}
	}

	if max, ok := options["max"]; ok {
		code += fmt.Sprintf(maxTemplate, fieldName, max, queryName, max)
	}

	if enumString, ok := options["enum"]; ok {
		enum := strings.Split(enumString, "|")
		param :=fmt.Sprintf( "[]%s{\"", filedType) + strings.Join(enum, "\",\"") + "\"}"
		code += fmt.Sprintf(checkEnumTemplate, param, fieldName, queryName)
	}

	return code
}

// createParamMethod - create Unpack and Validate methods for handler params
func createParamMethod(node *ast.File, typeName string) (string, string, error) {
	var unpackCode string
	var validateCode string
	for _, f := range node.Decls {
		g, ok := f.(*ast.GenDecl)
		if !ok { // Ignore not *ast.GenDecl items
			continue
		}
		for _, spec := range g.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok || currType.Name.Name != typeName { // Ignore not *ast.TypeSpec items
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok { // Get not *ast.StructType items
				continue
			}

			// Init unpack function
			var tmp bytes.Buffer
			err := unpackBaseTemplate.Execute(&tmp, typeNameData{TypeName: currType.Name.Name})
			if err != nil {
				log.Fatal(err)
			}
			unpackCode = tmp.String()

			// Init validate function
			tmp.Reset()
			err = validateTpl.Execute(&tmp, typeNameData{TypeName: currType.Name.Name})
			if err != nil {
				log.Fatal(err)
			}
			validateCode = tmp.String()

			for _, field := range currStruct.Fields.List {
				if field.Tag != nil {
					// Get tag option
					tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
					tagOptions := parseStructTag(tag.Get("apivalidator"))

					// Create unpack code for field
					unpack, err := createUnpackCode(field, tagOptions)
					if err != nil {
						log.Fatal(err)
					}
					unpackCode += unpack

					// Create validate code
					validateCode += createValidateCode(field, tagOptions)
				}
			}
			unpackCode += "\treturn nil\n}\n"
			validateCode += "\treturn nil\n}\n"
		}
	}
	return unpackCode, validateCode, nil
}

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := os.Create(os.Args[2])
	fmt.Fprintln(out, "package "+node.Name.Name+"\n")
	fmt.Fprintf(out, importStr)
	fmt.Fprintf(out, checkEnumFunc)

	for _, d := range node.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if ok {
			if fd.Doc != nil {
				for _, comment := range fd.Doc.List {
					if !strings.HasPrefix(comment.Text, generatorLabel) {
						// Ignore functions without label
						continue
					}
					config := new(handlerConfig)
					err := json.Unmarshal([]byte(comment.Text[14:]), config)
					if err != nil {
						log.Fatal(err)
					}

					// Create http wrapper for function
					wrapper, err := createWrapper(fd, config)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Fprintln(out, wrapper)

					// Create router for function
					err = createRoute(fd, config)
					if err != nil {
						log.Fatal(err)
					}

					// Create unpack and validate method for function parameter
					unpackCode, validateCode, err := createParamMethod(node, getParamType(fd))
					if err != nil {
						log.Fatal(err)
					}
					fmt.Fprintln(out, unpackCode)
					fmt.Fprintln(out, validateCode)
				}
			}
		}
	}

	// Create ServeHTTP methods
	for k, v := range serveHTTPRoutes {
		err := serveHTTPTpl.Execute(out, serveHTTPTemplateData{k, v})
		if err != nil {
			log.Fatal(err)
		}
	}
}
