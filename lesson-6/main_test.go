package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// CaseResponse
type CR map[string]interface{}

type Case struct {
	Method string // GET по-умолчанию в http.NewRequest если передали пустую строку
	Path   string
	Query  string
	Status int
	Result interface{}
	Body   interface{}
}

var (
	client = &http.Client{Timeout: time.Second}
)

func PrepareTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS items;`,

		`CREATE TABLE items (
  id int(11) NOT NULL AUTO_INCREMENT,
  title varchar(255) NOT NULL,
  description text NOT NULL,
  updated varchar(255) DEFAULT NULL,
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

		`INSERT INTO items (id, title, description, updated) VALUES
(1,	'database/sql',	'Рассказать про базы данных',	'rvasily'),
(2,	'memcache',	'Рассказать про мемкеш с примером использования',	NULL);`,

		`DROP TABLE IF EXISTS users;`,

		`CREATE TABLE users (
			user_id int(11) NOT NULL AUTO_INCREMENT,
  login varchar(255) NOT NULL,
  password varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  info text NOT NULL,
  updated varchar(255) DEFAULT NULL,
  PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

		`INSERT INTO users (user_id, login, password, email, info, updated) VALUES(1,	'rvasily',	'love',	'rvasily@example.com',	'none',	NULL);`,
	}

	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func CleanupTestApis(db *sql.DB) {
	qs := []string{
		`DROP TABLE IF EXISTS items;`,
		`DROP TABLE IF EXISTS users;`,
	}
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}

func TestApis(t *testing.T) {
	db, err := sql.Open("mysql", DSN)
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	//PrepareTestApis(db)

	// возможно вам будет удобно закомментировать это чтобы смотреть результат после теста
	//defer CleanupTestApis(db)

	handler, err := NewDbExplorer(db)
	if err != nil {
		panic(err)
	}

	ts := httptest.NewServer(handler)

	cases := []Case{
		Case{
			Path: "/", // список таблиц
			Result: CR{
				"response": CR{
					"tables": []string{"items", "users"},
				},
			},
		},
		Case{
			Path:   "/unknown_table",
			Status: http.StatusNotFound,
			Result: CR{
				"error": "unknown table",
			},
		},
		Case{
			Path: "/items",
			Result: CR{
				"response": CR{
					"records": []CR{
						CR{
							"id":          1,
							"title":       "database/sql",
							"description": "Рассказать про базы данных",
							"updated":     "rvasily",
						},
						CR{
							"id":          2,
							"title":       "memcache",
							"description": "Рассказать про мемкеш с примером использования",
							"updated":     nil,
						},
					},
				},
			},
		},
		Case{
			Path:  "/items",
			Query: "limit=1",
			Result: CR{
				"response": CR{
					"records": []CR{
						CR{
							"id":          1,
							"title":       "database/sql",
							"description": "Рассказать про базы данных",
							"updated":     "rvasily",
						},
					},
				},
			},
		},
		Case{
			Path:  "/items",
			Query: "limit=1&offset=1",
			Result: CR{
				"response": CR{
					"records": []CR{
						CR{
							"id":          2,
							"title":       "memcache",
							"description": "Рассказать про мемкеш с примером использования",
							"updated":     nil,
						},
					},
				},
			},
		},
		// тут тоже возможна sql-инъекция
		// если пришло не число на вход - берём дефолтное значене для лимита-оффсета
		//Case{
		//	Path:  "/users",
		//	Query: "limit=1'&offset=1\"",
		//	Result: CR{
		//		"response": CR{
		//			"records": []CR{
		//				CR{
		//					"user_id":  1,
		//					"login":    "rvasily",
		//					"password": "love",
		//					"email":    "rvasily@example.com",
		//					"info":     "try update",
		//					"updated":  "now",
		//				},
		//				CR{
		//					"user_id":  2,
		//					"login":    "qwerty'",
		//					"password": "love\"",
		//					"email":    "",
		//					"info":     "",
		//					"updated":  nil,
		//				},
		//			},
		//		},
		//	},
		//},
	}

	runCases(t, ts, db, cases)
}

func runCases(t *testing.T, ts *httptest.Server, db *sql.DB, cases []Case) {
	for idx, item := range cases {
		var (
			err      error
			result   interface{}
			expected interface{}
			req      *http.Request
		)

		caseName := fmt.Sprintf("case %d: [%s] %s %s", idx, item.Method, item.Path, item.Query)

		// если у вас случилась это ошибка - значит вы не делаете где-то rows.Close и у вас текут соединения с базой
		// если такое случилось на первом тесте - значит вы не закрываете коннект где-то при иницаилизации в NewDbExplorer
		if db.Stats().OpenConnections != 1 {
			t.Fatalf("[%s] you have %d open connections, must be 1", caseName, db.Stats().OpenConnections)
		}

		if item.Method == "" || item.Method == http.MethodGet {
			req, err = http.NewRequest(item.Method, ts.URL+item.Path+"?"+item.Query, nil)
		} else {
			data, err := json.Marshal(item.Body)
			if err != nil {
				panic(err)
			}
			reqBody := bytes.NewReader(data)
			req, err = http.NewRequest(item.Method, ts.URL+item.Path, reqBody)
			req.Header.Add("Content-Type", "application/json")
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("[%s] request error: %v", caseName, err)
			continue
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		// fmt.Printf("[%s] body: %s\n", caseName, string(body))
		if item.Status == 0 {
			item.Status = http.StatusOK
		}

		if resp.StatusCode != item.Status {
			t.Fatalf("[%s] expected http status %v, got %v", caseName, item.Status, resp.StatusCode)
			continue
		}

		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("[%s] cant unpack json: %v", caseName, err)
			continue
		}

		// reflect.DeepEqual не работает если нам приходят разные типы
		// а там приходят разные типы (string VS interface{}) по сравнению с тем что в ожидаемом результате
		// этот маленький грязный хак конвертит данные сначала в json, а потом обратно в interface - получаем совместимые результаты
		// не используйте это в продакшен-коде - надо явно писать что ожидается интерфейс или использовать другой подход с точным форматом ответа
		data, err := json.Marshal(item.Result)
		json.Unmarshal(data, &expected)

		if !reflect.DeepEqual(result, expected) {
			t.Fatalf("[%s] results not match\nGot : %#v\nWant: %#v", caseName, result, expected)
			continue
		}
	}

}
