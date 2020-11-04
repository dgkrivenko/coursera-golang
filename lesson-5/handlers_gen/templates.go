package main

import (
	"text/template"
)

var (
	serveHTTPTpl = template.Must(template.New("serveHTTPBase").Parse(`
func (h *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
{{.Router}}
	default:
		// 404
	}
}
`))

//	wrapperTpl = template.Must(template.New("wrapperTpl").Parse(`
//func (h *{{.structName}} ) wrapper{{.funcName}}(w http.ResponseWriter, r *http.Request) {
//	// заполнение структуры params
//	// валидирование параметров
//	res, err := h.{{.funcName}(ctx, params)
//	// прочие обработки
//}
//`))

	routeTpl = template.Must(template.New("routeTpl").Parse(`	case "{{.URL}}":
		h.wrapper{{.FuncName}}(w, r)
`))
	)