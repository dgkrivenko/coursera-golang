package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *MyApi) wrapperProfile(w http.ResponseWriter, r *http.Request) {
	var respBody []byte

	params := ProfileParams{}
	
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
}

func (h *MyApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	var respBody []byte

	params := CreateParams{}
	
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
}

func (h *OtherApi) wrapperCreate(w http.ResponseWriter, r *http.Request) {
	var respBody []byte

	params := OtherCreateParams{}
	
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
