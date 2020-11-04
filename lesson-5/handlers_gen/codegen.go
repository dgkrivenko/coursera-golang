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
	"strings"
)

const generatorLabel = "// apigen:api"

var serveHTTPRoutes = map[string]string{}

type HandlerConfig struct {
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
}

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

func getParamType(fd *ast.FuncDecl) string {
	i, ok := fd.Type.Params.List[1].Type.(*ast.Ident)
	if ok {
		return i.Name
	}
	return ""
}

func createWrapper(fd *ast.FuncDecl, config *HandlerConfig) (string, error) {
	rt := getReceiverType(fd)
	pt := getParamType(fd)
	var tmp bytes.Buffer
	err := wrapperTpl.Execute(&tmp, wrapperTemplateData{rt, pt, fd.Name.Name})
	if err != nil {
		return "", err
	}
	return tmp.String(), nil
}

func createRoute(fd *ast.FuncDecl, config *HandlerConfig) error {
	typeName := getReceiverType(fd)
	var tmp bytes.Buffer
	err := routeTpl.Execute(&tmp, routeTemplateData{config.URL, fd.Name.Name})
	if err != nil {
		return err
	}
	serveHTTPRoutes[typeName] += tmp.String()
	return nil
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

	for _, d := range node.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fd.Doc != nil {
			for _, comment := range fd.Doc.List {
				if !strings.HasPrefix(comment.Text, generatorLabel) {
					continue
				}
				config := new(HandlerConfig)
				err := json.Unmarshal([]byte(comment.Text[14:]), config)
				if err != nil {
					log.Fatal(err)
				}

				wrapper, err := createWrapper(fd, config)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Fprintln(out, wrapper)

				err = createRoute(fd, config)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	for k, v := range serveHTTPRoutes {
		err := serveHTTPTpl.Execute(out, serveHTTPTemplateData{k, v})
		if err != nil {
			log.Fatal(err)
		}
	}
}
