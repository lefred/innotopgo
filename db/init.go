package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(mysql_uri string) *sql.DB {
	fmt.Println(mysql_uri)
	db, err := sql.Open("mysql", mysql_uri)
	if err != nil {
		panic(err)
	}
	return db
}

func Query(db *sql.DB, stmt string) *sql.Rows {
	rows, err := db.Query(stmt)
	if err != nil {
		panic(err)
	}
	return rows
}

func GetData(rows *sql.Rows) ([]string, [][]string, error) {
	var result [][]string
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]interface{}, len(cols))
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, nil, err
		}
		var resultRow []string
		for _, col := range vals {
			if col == nil {
				resultRow = append(resultRow, "NULL")
			} else {
				resultRow = append(resultRow, fmt.Sprintf("%v", col))
			}
		}
		result = append(result, resultRow)
	}
	return cols, result, nil
}
