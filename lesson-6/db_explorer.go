package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// DBExplorer - database manager
// db - database connection
// tables - information about tables it's columns
type DBExplorer struct {
	db        *sql.DB
	tables    map[string]map[string]tableInfo
	singleURL *regexp.Regexp
	paramURL  *regexp.Regexp
}

// tableInfo - table information
type tableInfo struct {
	ColumnName sql.NullString
	ColumnType sql.NullString
	Collation  sql.NullString
	Null       sql.NullString
	Key        sql.NullString
	Default    sql.NullString
	Extra      sql.NullString
	Privileges sql.NullString
	Comment    sql.NullString
}

// New - Create new DBExplorer object
func New(db *sql.DB) (*DBExplorer, error) {
	showTables := "SHOW TABLES;"
	showColumns := "SHOW FULL COLUMNS FROM %s;"

	instance := new(DBExplorer)
	instance.db = db
	instance.tables = map[string]map[string]tableInfo{}

	// Get tables list from database
	rows, err := db.Query(showTables)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		instance.tables[name] = map[string]tableInfo{}
	}

	// Get tables info from database
	for k, _ := range instance.tables {
		rows, err := instance.db.Query(fmt.Sprintf(showColumns, k))
		if err != nil {
			return nil, err
		}

		data := tableInfo{}
		for rows.Next() {
			err := rows.Scan(&data.ColumnName, &data.ColumnType, &data.Collation, &data.Null, &data.Key,
				&data.Default, &data.Extra, &data.Privileges, &data.Comment)
			if err != nil {
				return nil, err
			}
			instance.tables[k][data.ColumnName.String] = data
		}
	}

	// Compile url regexp
	instance.singleURL, err = regexp.Compile(`^/.+$`)
	if err != nil {
		return nil, err
	}
	instance.paramURL, err = regexp.Compile(`^/.+/.`)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (e *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/" {
		e.mainPageHandler(w, r)
		return
	}

	if e.paramURL.MatchString(r.URL.Path) {
		if r.Method == http.MethodGet {
			e.getRecordHandler(w, r)
		} else if r.Method == http.MethodPost {
			e.updateRecordHandler(w, r)
		} else if r.Method == http.MethodDelete{
			e.deleteRecordHandler(w, r)
		} else {
			fmt.Println("Method not allowed")
		}
	} else if e.singleURL.MatchString(r.URL.Path) {

		if r.Method == http.MethodGet {
			e.recordListHandler(w, r)
		} else if r.Method == http.MethodPut {
			e.createRecordHandler(w, r)
		} else {
			fmt.Println("Method not allowed")
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte{})
	}
}

func NewDbExplorer(db *sql.DB) (*DBExplorer, error) {
	instance, err := New(db)
	if err != nil {
		panic(err)
	}

	return instance, nil
}

// mainPageHandler - returns all available tables
func (e *DBExplorer) mainPageHandler(w http.ResponseWriter, r *http.Request) {
	tables := map[string][]string{
		"tables": make([]string, 0, len(e.tables)),
	}
	for k, _ := range e.tables {
		tables["tables"] = append(tables["tables"], k)
	}

	response := map[string]interface{}{"response": tables}
	data, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_, err = w.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *DBExplorer) recordListHandler(w http.ResponseWriter, r *http.Request) {
	var limit = 5
	var offset int

	// if the table does not exist return an error
	table := strings.Split(r.URL.Path, "/")[1]
	if _, ok := e.tables[table]; !ok {
		w.WriteHeader(http.StatusNotFound)
		resp := map[string]string{"error": "unknown table"}
		data, _ := json.Marshal(resp)
		_, err := w.Write(data)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	// Set limit value
	v := r.URL.Query().Get("limit")
	if v != "" {
		limitVal, err := strconv.Atoi(v)
		if err == nil {
			limit = limitVal
		}
	}

	// Set offset value
	v = r.URL.Query().Get("offset")
	if v != "" {
		offsetVal, err := strconv.Atoi(v)
		if err == nil {
			offset = offsetVal
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", table, limit, offset)
	rows, err := e.db.Query(query)
	if err != nil {
		errorHandler(err, w)
		return
	}

	resp := map[string]map[string][]interface{}{
		"response": {
			"records": make([]interface{}, 0),
		},
	}

	for rows.Next() {
		colsTypes, err := rows.ColumnTypes()
		if err != nil {
			errorHandler(err, w)
			return
		}

		// Create buffer for scan operation
		columnsNum := len(colsTypes)
		var buf = make([]interface{}, 0, columnsNum)

		// Scan data
		for i := 0; i < columnsNum; i++ {
			buf = append(buf, createScanBuffer(colsTypes[i].DatabaseTypeName()))
		}
		err = rows.Scan(buf...)
		if err != nil {
			errorHandler(err, w)
			return
		}

		record := map[string]interface{}{}
		for i, v := range colsTypes{
			record[v.Name()] = buf[i]
		}
		resp["response"]["records"] = append(resp["response"]["records"], record)
	}
	data, _ := json.Marshal(resp)
	_, err = w.Write(data)
	if err != nil {
		errorHandler(err, w)
		return
	}

}

func (e *DBExplorer) getRecordHandler(w http.ResponseWriter, r *http.Request) {

}

func (e *DBExplorer) createRecordHandler(w http.ResponseWriter, r *http.Request) {

}

func (e *DBExplorer) updateRecordHandler(w http.ResponseWriter, r *http.Request) {

}

func (e *DBExplorer) deleteRecordHandler(w http.ResponseWriter, r *http.Request) {

}

func createScanBuffer(typeName string) interface{} {
	switch typeName {
	case "INT":
		return new(int)
	case "VARCHAR":
		return new(*string)
	case "TEXT":
		return new(*string)
	}
	return nil
}

func errorHandler(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte{})
	if err != nil {
		log.Fatal(err)
	}
}