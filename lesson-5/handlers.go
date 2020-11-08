package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	
	
	var respBody []byte
	params := ProfileParams{}
	err := params.Unpack(r)
	if err != nil {
		log.Fatal(err)
	}
	res, err := h.Profile(r.Context(), params)
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

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Fatal(err)
	}
}
func (in *ProfileParams) Unpack(r *http.Request) error {

	// Login unpack
	var valLogin string
	keysLogin, ok := r.URL.Query()["login"]
	if !ok || len(keysLogin[0]) < 1{
		valLogin = ""
	} else {
		valLogin = keysLogin[0]
	}
	in.Login = valLogin
	return nil
}


func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
    }
	authValue := r.Header.Get("X-Auth")
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
	}
	var respBody []byte
	params := CreateParams{}
	err := params.Unpack(r)
	if err != nil {
		log.Fatal(err)
	}
	res, err := h.Create(r.Context(), params)
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

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Fatal(err)
	}
}
func (in *CreateParams) Unpack(r *http.Request) error {

	// Login unpack
	var valLogin string
	keysLogin, ok := r.URL.Query()["login"]
	if !ok || len(keysLogin[0]) < 1{
		valLogin = ""
	} else {
		valLogin = keysLogin[0]
	}
	in.Login = valLogin

	// Name unpack
	var valName string
	keysName, ok := r.URL.Query()["full_name"]
	if !ok || len(keysName[0]) < 1{
		valName = ""
	} else {
		valName = keysName[0]
	}
	in.Name = valName

	// Status unpack
	var valStatus string
	keysStatus, ok := r.URL.Query()["status"]
	if !ok || len(keysStatus[0]) < 1{
		valStatus = "user"
	} else {
		valStatus = keysStatus[0]
	}
	in.Status = valStatus

	// Age unpack
	var valAge int
	keysAge, ok := r.URL.Query()["age"]
	if !ok || len(keysAge[0]) < 1{
		valAge = 0
	} else {
		valAge, _ = strconv.Atoi(keysAge[0])
		if valAge == 0 {
			valAge = 0
		}
	}
	in.Age = valAge
	return nil
}


func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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
    }
	authValue := r.Header.Get("X-Auth")
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
	}
	var respBody []byte
	params := OtherCreateParams{}
	err := params.Unpack(r)
	if err != nil {
		log.Fatal(err)
	}
	res, err := h.Create(r.Context(), params)
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

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(respBody)
	if err != nil {
		log.Fatal(err)
	}
}
func (in *OtherCreateParams) Unpack(r *http.Request) error {

	// Username unpack
	var valUsername string
	keysUsername, ok := r.URL.Query()["username"]
	if !ok || len(keysUsername[0]) < 1{
		valUsername = ""
	} else {
		valUsername = keysUsername[0]
	}
	in.Username = valUsername

	// Name unpack
	var valName string
	keysName, ok := r.URL.Query()["account_name"]
	if !ok || len(keysName[0]) < 1{
		valName = ""
	} else {
		valName = keysName[0]
	}
	in.Name = valName

	// Class unpack
	var valClass string
	keysClass, ok := r.URL.Query()["class"]
	if !ok || len(keysClass[0]) < 1{
		valClass = "warrior"
	} else {
		valClass = keysClass[0]
	}
	in.Class = valClass

	// Level unpack
	var valLevel int
	keysLevel, ok := r.URL.Query()["level"]
	if !ok || len(keysLevel[0]) < 1{
		valLevel = 0
	} else {
		valLevel, _ = strconv.Atoi(keysLevel[0])
		if valLevel == 0 {
			valLevel = 0
		}
	}
	in.Level = valLevel
	return nil
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.wrapperProfile(w, r)
	case "/user/create":
		h.wrapperCreate(w, r)

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
func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.wrapperCreate(w, r)

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
