package innotop

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/widgets/text"
)

func GetQueryByThreadId(mydb *sql.DB, thread_id string) (string, string, error) {
	stmt := fmt.Sprintf("select db, current_statement from sys.x$processlist where thd_id=%s;", thread_id)
	rows := db.Query(mydb, stmt)
	_, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}
	var query_text string
	var query_db string
	for _, row := range data {
		query_db = row[0]
		query_text = row[1]
	}

	return query_db, query_text, err
}

func GetExplain(mydb *sql.DB, explain_type string, query_db string, query_test string) ([]string, [][]string, error) {
	if len(query_db) > 0 {
		_, err := mydb.Exec("USE " + query_db)
		if err != nil {
			panic(err)
		}
	}
	if explain_type == "NORMAL" {
		explain_type = ""
	}
	stmt := fmt.Sprintf("EXPLAIN %s %s", explain_type, query_test)
	rows := db.Query(mydb, stmt)

	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}

func DisplayExplain(mydb *sql.DB, c *container.Container, main_window *text.Text, thread_id string, explain_type string) error {
	var line string

	query_db, query_text, err := GetQueryByThreadId(mydb, thread_id)
	if err != nil {
		panic(err)
	}
	cols, data, err := GetExplain(mydb, explain_type, query_db, query_text)

	for _, row := range data {
		i := 0
		for _, col := range cols {
			if len(cols) > 1 {
				line = fmt.Sprintf("%15v: %-v\n", col, row[i])
			} else {
				line = fmt.Sprintf("%-v\n", row[i])
			}
			i++
			main_window.Write(line)
		}
		if explain_type == "NORMAL" {
			main_window.Write(strings.Repeat("*", 100) + "\n")
		}
	}
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("EXPLAIN (<-- <Backspace> to return  -  <Space> to change EXPLAIN FORMAT)"))

	return err
}
