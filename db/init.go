package db

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func Connect(mysql_uri string) *sql.DB {
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

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil, err
	}
	vals := make([]interface{}, len(cols))
	for i, _ := range cols {
		vals[i] = new(sql.RawBytes)
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			return nil, nil, err
		}
		var resultRow []string
		for i, col := range vals {
			var value string
			if col == nil {
				value = "NULL"
			} else {
				switch colTypes[i].DatabaseTypeName() {
				case "VARCHAR", "CHAR", "TEXT":
					value = fmt.Sprintf("%s", col)
				case "BIGINT":
					value = fmt.Sprintf("%s", col)
				case "INT":
					value = fmt.Sprintf("%d", col)
				case "DECIMAL":
					value = fmt.Sprintf("%s", col)
				default:
					value = fmt.Sprintf("%s:%T", cols[i], col)
				}
			}
			value = strings.Replace(value, "&", "", 1)
			resultRow = append(resultRow, value)
		}
		result = append(result, resultRow)
	}
	return cols, result, nil
}

func GetServerInfo(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `select @@version_comment, @@version, @@hostname, @@port`
	rows := Query(mydb, stmt)
	cols, data, err := GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}
