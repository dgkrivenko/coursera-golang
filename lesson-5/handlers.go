package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)
func checkEnum(enum []string, value string) bool {
	for _, v := range enum {
		if v == value {
			return true
		}
	}
	return false
}
func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	
	
	var respBody []byte
	params := ProfileParams{}
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
		res, err := h.Profile(r.Context(), params)
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
}
func (in *ProfileParams) Unpack(r *http.Request) error {
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
	}
	
	// Login unpack
	var valLogin string
	keysLogin, ok := values["login"]
	if !ok || len(keysLogin[0]) < 1{
		valLogin = ""
	} else {
		valLogin = keysLogin[0]
	}
	in.Login = valLogin
	return nil
}

func (in ProfileParams) Validate() error {
	if in.Login == "" { return fmt.Errorf("login must me not empty")}
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
		res, err := h.Create(r.Context(), params)
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
}
func (in *CreateParams) Unpack(r *http.Request) error {
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
	}
	
	// Login unpack
	var valLogin string
	keysLogin, ok := values["login"]
	if !ok || len(keysLogin[0]) < 1{
		valLogin = ""
	} else {
		valLogin = keysLogin[0]
	}
	in.Login = valLogin

	
	// Name unpack
	var valName string
	keysName, ok := values["full_name"]
	if !ok || len(keysName[0]) < 1{
		valName = ""
	} else {
		valName = keysName[0]
	}
	in.Name = valName

	
	// Status unpack
	var valStatus string
	keysStatus, ok := values["status"]
	if !ok || len(keysStatus[0]) < 1{
		valStatus = "user"
	} else {
		valStatus = keysStatus[0]
	}
	in.Status = valStatus

	// Age unpack
	var valAge int
	var err error
	keysAge, ok := values["age"]
	if !ok || len(keysAge[0]) < 1{
		valAge = 0
	} else {
		valAge, err = strconv.Atoi(keysAge[0])
		if err != nil {
			return fmt.Errorf("age must be int")
		}
		if valAge == 0 {
			valAge = 0
		}
	}
	in.Age = valAge
	return nil
}

func (in CreateParams) Validate() error {
	if in.Login == "" { return fmt.Errorf("login must me not empty")}
	if len(in.Login) < 10 { return fmt.Errorf("login len must be >= 10")}
	enumValues := []string{"user","moderator","admin"}
	if !checkEnum(enumValues, in.Status) {
		errorMsg := "status must be one of " + "[" + strings.Join(enumValues, ", ") + "]"
		return fmt.Errorf(errorMsg)
	}
	if in.Age < 0 { return fmt.Errorf("age must be >= 0")}
	if in.Age > 128 { return fmt.Errorf("age must be <= 128")}
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
		res, err := h.Create(r.Context(), params)
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
}
func (in *OtherCreateParams) Unpack(r *http.Request) error {
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
	}
	
	// Username unpack
	var valUsername string
	keysUsername, ok := values["username"]
	if !ok || len(keysUsername[0]) < 1{
		valUsername = ""
	} else {
		valUsername = keysUsername[0]
	}
	in.Username = valUsername

	
	// Name unpack
	var valName string
	keysName, ok := values["account_name"]
	if !ok || len(keysName[0]) < 1{
		valName = ""
	} else {
		valName = keysName[0]
	}
	in.Name = valName

	
	// Class unpack
	var valClass string
	keysClass, ok := values["class"]
	if !ok || len(keysClass[0]) < 1{
		valClass = "warrior"
	} else {
		valClass = keysClass[0]
	}
	in.Class = valClass

	// Level unpack
	var valLevel int
	var err error
	keysLevel, ok := values["level"]
	if !ok || len(keysLevel[0]) < 1{
		valLevel = 0
	} else {
		valLevel, err = strconv.Atoi(keysLevel[0])
		if err != nil {
			return fmt.Errorf("level must be int")
		}
		if valLevel == 0 {
			valLevel = 0
		}
	}
	in.Level = valLevel
	return nil
}

func (in OtherCreateParams) Validate() error {
	if in.Username == "" { return fmt.Errorf("username must me not empty")}
	if len(in.Username) < 3 { return fmt.Errorf("username len must be >= 3")}
	enumValues := []string{"warrior","sorcerer","rouge"}
	if !checkEnum(enumValues, in.Class) {
		errorMsg := "class must be one of " + "[" + strings.Join(enumValues, ", ") + "]"
		return fmt.Errorf(errorMsg)
	}
	if in.Level < 1 { return fmt.Errorf("level must be >= 1")}
	if in.Level > 50 { return fmt.Errorf("level must be <= 50")}
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
