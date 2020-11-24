package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	tables    map[string][]tableInfo
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
	instance.tables = map[string][]tableInfo{}

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
		instance.tables[name] = []tableInfo{}
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
			instance.tables[k] = append(instance.tables[k], data)
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

// ServeHTTP - implements http.Handler interface
func (e *DBExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path == "/" {
		e.mainPageHandler(w, r)
		return
	}

	// Router
	if e.paramURL.MatchString(r.URL.Path) {
		if r.Method == http.MethodGet {
			e.getRecordHandler(w, r)
		} else if r.Method == http.MethodPost {
			e.updateRecordHandler(w, r)
		} else if r.Method == http.MethodDelete {
			e.deleteRecordHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else if e.singleURL.MatchString(r.URL.Path) {

		if r.Method == http.MethodGet {
			e.recordListHandler(w, r)
		} else if r.Method == http.MethodPut {
			e.createRecordHandler(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte{})
	}
}

// NewDbExplorer - create new explorer instance
func NewDbExplorer(db *sql.DB) (*DBExplorer, error) {
	instance, err := New(db)
	if err != nil {
		panic(err)
	}
	return instance, nil
}

// errorResponse - return error response to client
func errorResponse(code int, message string, w http.ResponseWriter) {
	w.WriteHeader(code)
	if message == "" {
		_, err := w.Write([]byte{})
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	data := map[string]interface{}{
		"error": message,
	}
	resp, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	_, err = w.Write(resp)
	if err != nil {
		log.Fatal(err)
	}
}

// createScanBuffer - create slice with variables for scan sql operation result
func createScanBuffer(rows *sql.Rows) ([]interface{}, []*sql.ColumnType, error) {
	colsTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	// Create buffer for scan operation
	columnsNum := len(colsTypes)
	var buf = make([]interface{}, 0, columnsNum)

	// Scan data
	for i := 0; i < columnsNum; i++ {
		buf = append(buf, createBuffer(colsTypes[i].DatabaseTypeName()))
	}
	return buf, colsTypes, nil
}

// createBuffer - create buffer for single variable
func createBuffer(typeName string) interface{} {
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

// errorHandler - handle fatal erros
func errorHandler(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, err = w.Write([]byte{})
	if err != nil {
		log.Fatal(err)
	}
}

// getPKFieldName - get pk filed name
func (e *DBExplorer) getPKFieldName(table string) string {
	for _, v := range e.tables[table] {
		if v.Key.String == "PRI" {
			return v.ColumnName.String
		}
	}
	return ""
}

// validateTypes - validate type before insert
func (e *DBExplorer) validateTypes(table string, data map[string]interface{}) error {
	columns := e.tables[table]
	for _, c := range columns {
		v, ok := data[c.ColumnName.String]
		if !ok {
			continue
		}
		if c.Null.String == "YES" && v == nil {
			continue
		}

		switch c.ColumnType.String {
		case "int(11)":
			_, ok := v.(float64)
			if !ok {
				return fmt.Errorf("field %s have invalid type", c.ColumnName.String)
			}
		case "varchar(255)":
			_, ok := v.(string)
			if !ok {
				return fmt.Errorf("field %s have invalid type", c.ColumnName.String)
			}
		case "text":
			_, ok := v.(string)
			if !ok {
				return fmt.Errorf("field %s have invalid type", c.ColumnName.String)
			}
		case "float":
			_, ok := v.(float64)
			if !ok {
				return fmt.Errorf("field %s have invalid type", c.ColumnName.String)
			}
		}
	}
	return nil
}

// getDefaultValue - get default value for column
func getDefaultValue(info tableInfo) interface{} {
	if !info.Default.Valid && info.Null.String == "YES" {
		return nil
	} else if !info.Default.Valid {
		switch info.ColumnType.String {
		case "int(11)":
			return float64(0)
		case "varchar(255)":
			return ""
		case "text":
			return ""
		case "float":
			return float64(0)
		}
	}
	return info.Default.String
}

// filterColumns - delete unknown fields
func (e *DBExplorer) filterColumns(table string, data map[string]interface{}) {
DataLoop:
	for k, _ := range data {
		for _, c := range e.tables[table] {
			if c.ColumnName.String == k {
				continue DataLoop
			}
		}
		delete(data, k)
	}
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

// recordListHandler - return all records from table
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
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	defer rows.Close()

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

		// Create a buffer to read data
		buf, colsTypes, err := createScanBuffer(rows)
		if err != nil {
			errorResponse(http.StatusInternalServerError, "", w)
			return
		}

		err = rows.Scan(buf...)
		if err != nil {
			errorHandler(err, w)
			return
		}

		record := map[string]interface{}{}
		for i, v := range colsTypes {
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

// getRecordHandler - return detail record info
func (e *DBExplorer) getRecordHandler(w http.ResponseWriter, r *http.Request) {
	// if the table does not exist return an error
	table := strings.Split(r.URL.Path, "/")[1]
	if _, ok := e.tables[table]; !ok {
		errorResponse(http.StatusNotFound, "unknown table", w)
		return
	}

	// Get record id
	recordID := strings.Split(r.URL.Path, "/")[2]
	id, err := strconv.Atoi(recordID)
	if err != nil {
		errorResponse(http.StatusBadRequest, "field id title have invalid type", w)
		return
	}

	// Execute sql-query
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, e.getPKFieldName(table))
	rows, err := e.db.Query(query, id)
	if err != nil {
		errorHandler(err, w)
		return
	}
	defer rows.Close()

	// Create a buffer to read data
	buf, colsTypes, err := createScanBuffer(rows)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			errorResponse(http.StatusInternalServerError, "", w)
			return
		}
		errorResponse(http.StatusNotFound, "record not found", w)
		return
	}
	err = rows.Scan(buf...)
	if err != nil {
		errorHandler(err, w)
		return
	}

	record := map[string]interface{}{}
	for i, v := range colsTypes {
		record[v.Name()] = buf[i]
	}

	resp := map[string]map[string]interface{}{
		"response": {
			"record": record,
		},
	}

	data, _ := json.Marshal(resp)
	_, err = w.Write(data)
	if err != nil {
		errorHandler(err, w)
		return
	}
}

// createRecordHandler - create new record
func (e *DBExplorer) createRecordHandler(w http.ResponseWriter, r *http.Request) {
	// if the table does not exist return an error
	table := strings.Split(r.URL.Path, "/")[1]
	if _, ok := e.tables[table]; !ok {
		errorResponse(http.StatusNotFound, "unknown table", w)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	err = e.validateTypes(table, data)
	if err != nil {
		errorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	e.filterColumns(table, data)

	pk := e.getPKFieldName(table)

	query := "INSERT INTO %s (%s) VALUES (%s)"
	filedNames := make([]string, 0, len(data))
	filedValues := make([]interface{}, 0, len(data))
	esc := make([]string, 0, len(data))

	for _, c := range e.tables[table] {
		if c.ColumnName.String == pk {
			continue
		}
		filedNames = append(filedNames, c.ColumnName.String)
		esc = append(esc, "?")
		v, ok := data[c.ColumnName.String]
		if ok {
			filedValues = append(filedValues, v)
		} else {
			filedValues = append(filedValues, getDefaultValue(c))
		}
	}

	fields := strings.Join(filedNames, ", ")
	valuesEsc := strings.Join(esc, ",")
	result, err := e.db.Exec(fmt.Sprintf(query, table, fields, valuesEsc), filedValues...)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	response := map[string]interface{}{
		"response": map[string]int64{
			pk: id,
		},
	}

	resp, _ := json.Marshal(response)
	_, err = w.Write(resp)
	if err != nil {
		errorHandler(err, w)
		return
	}
}

// updateRecordHandler - update record info
func (e *DBExplorer) updateRecordHandler(w http.ResponseWriter, r *http.Request) {
	// if the table does not exist return an error
	table := strings.Split(r.URL.Path, "/")[1]
	if _, ok := e.tables[table]; !ok {
		errorResponse(http.StatusNotFound, "unknown table", w)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	data := map[string]interface{}{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}

	pk := e.getPKFieldName(table)

	// Get record id
	recordID := strings.Split(r.URL.Path, "/")[2]
	id, err := strconv.Atoi(recordID)
	if err != nil {
		msg := fmt.Sprintf("field %d title have invalid type", id)
		errorResponse(http.StatusBadRequest, msg, w)
		return
	}

	err = e.validateTypes(table, data)
	if err != nil {
		errorResponse(http.StatusBadRequest, err.Error(), w)
		return
	}

	query := "UPDATE %s SET %s WHERE %s=?"
	updates := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))
	for k, v := range data {
		if k == pk { // Primary key cannot be updated on an existing record
			msg := fmt.Sprintf("field %s have invalid type", pk)
			errorResponse(http.StatusBadRequest, msg, w)
			return
		}
		updates = append(updates, fmt.Sprintf("%s=?", k))
		values = append(values, v)
	}
	values = append(values, id)

	updateStr := strings.Join(updates, ", ")
	result, err := e.db.Exec(fmt.Sprintf(query, table, updateStr, pk), values...)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	response := map[string]interface{}{
		"response": map[string]int64{
			"updated": count,
		},
	}

	resp, _ := json.Marshal(response)
	_, err = w.Write(resp)
	if err != nil {
		errorHandler(err, w)
		return
	}
}

// deleteRecordHandler - delete record info
func (e *DBExplorer) deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	// if the table does not exist return an error
	table := strings.Split(r.URL.Path, "/")[1]
	if _, ok := e.tables[table]; !ok {
		errorResponse(http.StatusNotFound, "unknown table", w)
		return
	}

	// Get record id
	recordID := strings.Split(r.URL.Path, "/")[2]
	id, err := strconv.Atoi(recordID)
	if err != nil {
		errorResponse(http.StatusBadRequest, "field id title have invalid type", w)
		return
	}
	pk := e.getPKFieldName(table)

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, pk)
	result, err := e.db.Exec(query, id)
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		errorResponse(http.StatusInternalServerError, "", w)
		return
	}
	response := map[string]interface{}{
		"response": map[string]int64{
			"deleted": count,
		},
	}
	resp, _ := json.Marshal(response)
	_, err = w.Write(resp)
	if err != nil {
		errorHandler(err, w)
		return
	}
}