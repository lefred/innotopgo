package innotop

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/text"
)

func GetQueryByThreadId(mydb *sql.DB, thread_id string) (string, string, error) {
	stmt := fmt.Sprintf("select db, current_statement from sys.x$processlist where thd_id=%s;", thread_id)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return "", "", err
	}
	_, data, err := db.GetData(rows)
	if err != nil {
		return "", "", err
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
	mydb.SetMaxOpenConns(1)
	if len(query_db) > 0 {
		_, err := mydb.Exec("USE " + query_db)
		if err != nil {
			return nil, nil, err
		}
	}
	if explain_type == "NORMAL" {
		explain_type = ""
	}
	stmt := fmt.Sprintf("EXPLAIN %s %s", explain_type, query_test)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}

	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}
	mydb.SetMaxOpenConns(0)
	return cols, data, err
}

func DisplayExplain(mydb *sql.DB, c *container.Container, top_window *text.Text, main_window *text.Text, thread_id string, explain_type string) error {
	var line string
	var err error
	query_db, query_text, err := GetQueryByThreadId(mydb, thread_id)
	if err != nil {
		return err
	}
	if len(query_text) < 8 {
		// 8 is just an abitrary number
		top_window.Write("No Query", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
		main_window.Write("\n\nNothing to EXPLAIN...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
		err = nil
	} else {
		// print the query on top
		if explain_type == "NORMAL" {
			top_window.Reset()
			top_window.Write(query_text)
		}
		cols, data, err := GetExplain(mydb, explain_type, query_db, query_text)
		if err != nil {
			return err
		}
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
	}
	c.Update("dyn_top_container", container.SplitHorizontal(container.Top(
		container.Border(linestyle.Light),
		container.ID("top_container"),
		container.PlaceWidget(top_window),
	),
		container.Bottom(
			container.Border(linestyle.Light),
			container.ID("main_container"),
			container.PlaceWidget(main_window),
			container.FocusedColor(cell.ColorNumber(15)),
		), container.SplitFixed(10)))
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("EXPLAIN (<-- <Backspace> to return  -  <Space> to change EXPLAIN FORMAT)"))

	return err
}
