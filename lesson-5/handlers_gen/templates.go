package main

import (
	"text/template"
)

const importStr = `import (
	"encoding/json"
	"log"
	"net/http"
)
`

var (
	serveHTTPTpl = template.Must(template.New("serveHTTPBase").Parse(`func (h *{{.StructName}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
{{.Router}}
	default:
		response := map[string]string{"error": "unknown method"}
		body, err := json.Marshal(response)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, err = w.Write(body)
		if err != nil {
			log.Fatal(err)
		}
	}
}
`))

	wrapperTpl = template.Must(template.New("wrapperTpl").Parse(`
func (h *{{.ReceiverType}}) wrapper{{.FuncName}}(w http.ResponseWriter, r *http.Request) {
	var respBody []byte

	params := {{.ParamsType}}{}
	
	res, err := h.{{.FuncName}}(r.Context(), params)
	if err != nil {
		if apiError, ok := err.(ApiError); ok {
			resp := map[string]string{"error": apiError.Err.Error()}
			respBody, err = json.Marshal(resp)
			if err != nil {
				log.Fatal(err)
			}
			w.WriteHeader(apiError.HTTPStatus)
		}
	} else {
		respBody, err = json.Marshal(res)
		if err != nil {
			log.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Fatal(err)
	}
}`))

	routeTpl = template.Must(template.New("routeTpl").Parse(`	case "{{.URL}}":
		h.wrapper{{.FuncName}}(w, r)
`))
	)