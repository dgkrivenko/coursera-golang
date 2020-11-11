package main

import (
	"text/template"
)

const importStr = `import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)
`

func checkEnum(enum []string, value string) bool {
	for _, v := range enum {
		if v == value {
			return true
		}
	}
	return false
}

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
	{{.MethodCheck}}
	{{.AuthCheck}}
	var respBody []byte
	params := {{.ParamsType}}{}
	err := params.Unpack(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": err.Error()}
		respBody, err = json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBody)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = params.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := map[string]string{"error": err.Error()}
		respBody, err = json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		res, err := h.{{.FuncName}}(r.Context(), params)
		if err != nil {
			if apiError, ok := err.(ApiError); ok {
				w.WriteHeader(apiError.HTTPStatus)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp := map[string]string{"error": err.Error()}
			respBody, err = json.Marshal(resp)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			response := map[string]interface{}{
				"error": "",
				"response": res,
			}
			respBody, err = json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}
			w.WriteHeader(http.StatusOK)
		}
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

	intTpl = template.Must(template.New("intTpl").Parse(`
	// {{.FieldName}} unpack
	var val{{.FieldName}} int
	var err error
	keys{{.FieldName}}, ok := values["{{.QueryName}}"]
	if !ok || len(keys{{.FieldName}}[0]) < 1{
		val{{.FieldName}} = {{.DefaultVal}}
	} else {
		val{{.FieldName}}, err = strconv.Atoi(keys{{.FieldName}}[0])
		if err != nil {
			return fmt.Errorf("{{.QueryName}} must be int")
		}
		if val{{.FieldName}} == 0 {
			val{{.FieldName}} = {{.DefaultVal}}
		}
	}
	in.{{.FieldName}} = val{{.FieldName}}
`))
	strTpl = template.Must(template.New("strTpl").Parse(`
	
	// {{.FieldName}} unpack
	var val{{.FieldName}} string
	keys{{.FieldName}}, ok := values["{{.QueryName}}"]
	if !ok || len(keys{{.FieldName}}[0]) < 1{
		val{{.FieldName}} = {{.DefaultVal}}
	} else {
		val{{.FieldName}} = keys{{.FieldName}}[0]
	}
	in.{{.FieldName}} = val{{.FieldName}}
`))
	methodCheckTemplate = `if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusNotAcceptable)
		resp := map[string]string{"error": "bad method"}
		respBody, err := json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBody)
		if err != nil {
			log.Fatal(err)
		}
		return
    }`
	authCheckTemplate = `authValue := r.Header.Get("X-Auth")
	if authValue != "100500" {
		w.WriteHeader(http.StatusForbidden)
		resp := map[string]string{"error": "unauthorized"}
		respBody, err := json.Marshal(resp)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(respBody)
		if err != nil {
			log.Fatal(err)
		}
		return
	}`

	unpackBaseTemplate = template.Must(template.New("unpackBaseTemplate").Parse(`func (in *{{.TypeName}}) Unpack(r *http.Request) error {
	var values url.Values
	if r.Method == http.MethodGet {
		values = r.URL.Query()
	} else {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		values, err = url.ParseQuery(string(body))
		if err != nil {
			log.Fatal(err)
		}
	}`))
	validateTpl = template.Must(template.New("routeTpl").Parse(`func (in {{.TypeName}}) Validate() error {
`))
	checkEnumFunc = `func checkEnum(enum []string, value string) bool {
	for _, v := range enum {
		if v == value {
			return true
		}
	}
	return false
}`
	requiredTemplateInt = "\tif in.%s == 0 { return fmt.Errorf(\"%s must me not empty\")}\n"
	requiredTemplateString = "\tif in.%s == \"\" { return fmt.Errorf(\"%s must me not empty\")}\n"
	minTemplateInt = "\tif in.%s < %s { return fmt.Errorf(\"%s must be >= %s\")}\n"
	minTemplateString= "\tif len(in.%s) < %s { return fmt.Errorf(\"%s len must me >= %s\")}\n"
	maxTemplate = "\tif in.%s > %s { return fmt.Errorf(\"%s must be <= %s\")}\n"
	checkEnumTemplate = `	enumValues := %v
	if !checkEnum(enumValues, in.%s) {
		errorMsg := "%s must be one of " + "[" + strings.Join(enumValues, ", ") + "]"
		return fmt.Errorf(errorMsg)
	}
`
)
