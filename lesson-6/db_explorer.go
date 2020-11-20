package main

import (
	"database/sql"
	"fmt"
	"net/http"
)

// DBExplorer - database manager
// db - database connection
// tables - information about tables it's columns
type DBExplorer struct {
	db *sql.DB
	tables map[string]map[string]tableInfo
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

	return instance, nil
}

// mainPage - returns all available tables
func (e *DBExplorer) mainPage(w http.ResponseWriter, r *http.Request) {

}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	instance, err := New(db)
	if err != nil {
		panic(err)
	}

	fmt.Println(instance.tables)


	siteMux := http.NewServeMux()
	siteMux.HandleFunc("/", instance.mainPage)

	return siteMux, nil
}