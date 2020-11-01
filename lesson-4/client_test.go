package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Row struct {
	ID            string `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      string `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           string `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type Root struct {
	XMLName xml.Name `xml:"root"`
	Users []Row `xml:"row"`
}

func OrderAsc(arr []Row, orderField string) error {
	if orderField == "" || orderField == "Name" {
		sort.SliceStable(arr, func(i, j int) bool {
			return (arr[i].FirstName + arr[i].LastName)  < arr[j].FirstName + arr[j].LastName
		})
	} else if orderField == "Age" {
		sort.SliceStable(arr, func(i, j int) bool {
			return arr[i].Age < arr[j].Age
		})
	} else if orderField == "Id" {
		sort.SliceStable(arr, func(i, j int) bool {
			return arr[i].ID < arr[j].ID
		})
	} else {
		return fmt.Errorf("ErrorBadOrderField")
	}
	return nil
}

func OrderDesc(arr []Row, orderField string) error {
	if orderField == "" || orderField == "Name" {
		sort.SliceStable(arr, func(i, j int) bool {
			return (arr[i].FirstName + arr[i].LastName)  > arr[j].FirstName + arr[j].LastName
		})
	} else if orderField == "Age" {
		sort.SliceStable(arr, func(i, j int) bool {
			return arr[i].Age > arr[j].Age
		})
	} else if orderField == "Id" {
		sort.SliceStable(arr, func(i, j int) bool {
			return arr[i].ID > arr[j].ID
		})
	} else {
		return fmt.Errorf("ErrorBadOrderField")
	}
	return nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token != "test" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("wrong assess token"))
	}
	b, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	usersData := Root{}
	err = xml.Unmarshal(b, &usersData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy := r.URL.Query().Get("order_by")
	l := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(l)
	searchResult := make([]Row, 0)

	if query != "" {
		for _, user := range usersData.Users {
			isName := strings.Contains(user.FirstName + user.LastName, query)
			isAbout := strings.Contains(user.About, query)
			if isName || isAbout  {
				searchResult = append(searchResult, user)
			}
		}
	} else {
		searchResult = usersData.Users
	}

	if orderBy == "-1" {
		err = OrderAsc(searchResult, orderField)
		if err != nil {
			response := SearchErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			r, _ := json.Marshal(response)
			w.Write(r)
			return
		}
	} else if orderBy == "1" {
		err = OrderDesc(searchResult, orderField)
		if err != nil {
			response := SearchErrorResponse{
				Error: err.Error(),
			}
			w.WriteHeader(http.StatusBadRequest)
			r, _ := json.Marshal(response)
			w.Write(r)
			return
		}
	} else if orderBy != "0" {

		response := SearchErrorResponse{
			Error: "BadOrderBy",
		}
		w.WriteHeader(http.StatusBadRequest)
		r, _ := json.Marshal(response)
		w.Write(r)
		return
	}
	var data []User

	for _, user := range searchResult {
		id, _ := strconv.Atoi(user.ID)
		age, _ := strconv.Atoi(user.Age)
		result := User{
			Id:     id,
			Name:   user.FirstName + " " + user.LastName,
			Age:   age,
			About:  user.About,
			Gender: user.Gender,
		}
		data = append(data, result)
		if len(data) == limit {
			break
		}
	}

	response, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte{})
		return
	}

	w.Write(response)

}

func SearchServerTimeout(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	w.Write([]byte{})
}

func SearchServerUnknown(w http.ResponseWriter, r *http.Request)  {
	os.Exit(-1)
}

func SearchServerTimeoutInternalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte{})
}


var ts = httptest.NewServer(http.HandlerFunc(SearchServer))
var c = SearchClient{
	AccessToken: "test",
	URL: ts.URL,
}

func TestFindLimit(t *testing.T) {
	smallLimit := SearchRequest{
		Limit:      -1,
	}

	largeLimit := SearchRequest{
		Limit:      30,
	}

	normalLimit := SearchRequest{
		Limit: 10,
		Query: "test",
	}

	_, err := c.FindUsers(smallLimit)
	if err == nil || err.Error() != "limit must be > 0" {
		t.Errorf("small limit error")
	}

	_, err = c.FindUsers(largeLimit)
	if err != nil {
		t.Errorf("large limit error")
	}

	_, err = c.FindUsers(normalLimit)
	if err != nil {
		t.Errorf("normal limit error")
	}
}

func TestFindSuccess (t *testing.T) {
	r := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	}

	_, err := c.FindUsers(r)
	if err != nil {
		t.Errorf("success fail")
	}
}

func TestFindTimeout(t *testing.T) {
	tsTimeout := httptest.NewServer(http.HandlerFunc(SearchServerTimeout))
	clientTimeout := SearchClient{
		AccessToken: "test",
		URL: tsTimeout.URL,
	}
	r := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	}
	_, err := clientTimeout.FindUsers(r)
	if err == nil {
		t.Errorf("TestFindTimeout error")
	}
}

func TestFindInternalError(t *testing.T) {
	tsInternal := httptest.NewServer(http.HandlerFunc(SearchServerTimeoutInternalError))
	clientInternal := SearchClient{
		AccessToken: "test",
		URL: tsInternal.URL,
	}
	r := SearchRequest{
		Limit:      10,
		Offset:     0,
		Query:      "",
		OrderField: "",
		OrderBy:    0,
	}
	_, err := clientInternal.FindUsers(r)
	if err == nil {
		t.Errorf("TestFindInternalError error")
	}
}

func TestFindOffset(t *testing.T) {
	r := SearchRequest{Offset: -1}
	_, err := c.FindUsers(r)
	if err == nil {
		t.Errorf("TestOffset error")
	}
}

func TestFindUnknownError(t *testing.T) {
	tsUnknown := httptest.NewServer(http.HandlerFunc(SearchServerUnknown))
	clientUnknown := SearchClient{
		AccessToken: "test",
		URL: tsUnknown.URL + "12",
	}
	r := SearchRequest{}

	_, err := clientUnknown.FindUsers(r)
	if err == nil {
		t.Errorf("TestFindUnknownError fail")
	}
}

func TestBadRequest(t *testing.T) {
	badOrderField := SearchRequest{OrderField: "werwr", OrderBy: 1}
	_, err := c.FindUsers(badOrderField)
	if err == nil {
		t.Errorf("TestBadRequest badOrderField error")
	}

	unknownError := SearchRequest{OrderBy: 12}
	_, err = c.FindUsers(unknownError)
	if err == nil {
		t.Errorf("TestBadRequest unknownError error")
	}

}

func TestFindUnpack(t *testing.T) {
	tsUnpack1 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("[}"))
	}))
	tsUnpack2 := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("[}"))
	}))

	unpackError := SearchRequest{OrderBy: 12}

	clientUnknown := SearchClient{
		AccessToken: "test",
		URL: tsUnpack1.URL,
	}

	_, err := clientUnknown.FindUsers(unpackError)
	if err == nil {
		t.Errorf("TestFindUnpack1 error")
	}

	clientUnknown.URL = tsUnpack2.URL
	_, err = clientUnknown.FindUsers(unpackError)
	if err == nil {
		t.Errorf("unpackError error")
	}
}

func TestFindStatusCode(t *testing.T) {
	c.AccessToken = "test2"
	r := SearchRequest{}

	_, err := c.FindUsers(r)
	if err == nil {
		t.Errorf("TestFindStatusCode access fail")
	}

	c.AccessToken = "test"
}
