package innotop

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/widgets/text"
)

func GetStatus(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `select variable_name, variable_value from performance_schema.global_status`
	rows := db.Query(mydb, stmt)
	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}

func DisplayStatus(mydb *sql.DB, top_window *text.Text, prev_status map[string]string) (map[string]string, error) {
	var line string
	var real_qps int
	_, data, err := GetStatus(mydb)
	if err != nil {
		panic(err)
	}
	var status = make(map[string]string)
	for _, row := range data {
		status[row[0]] = row[1]
	}
	uptime_sec, _ := strconv.Atoi(status["Uptime"])
	queries, _ := strconv.Atoi(status["Queries"])

	if prev_status != nil {
		line = fmt.Sprintf("Uptime: %-10v", (time.Duration(uptime_sec) * time.Second))
		top_window.Write(line)
		top_window.Write("\n")
		prev_uptime_sec, _ := strconv.Atoi(prev_status["Uptime"])
		prev_queries, _ := strconv.Atoi(prev_status["Queries"])
		if (uptime_sec - prev_uptime_sec) < 1 {
			real_qps = 0
		} else {
			real_qps = (queries - prev_queries) / (uptime_sec - prev_uptime_sec)
		}
		line = fmt.Sprintf("   QPS: %-10v", (queries / uptime_sec))
		top_window.Write(line)
		line = fmt.Sprintf("real QPS: %v", real_qps)
		top_window.Write(line)

	} else {
		line = fmt.Sprintf("Uptime: %-10v", (time.Duration(uptime_sec) * time.Second))
		top_window.Write(line)
		top_window.Write("\n")

		line = fmt.Sprintf("   QPS: %-10v", (queries / uptime_sec))
		top_window.Write(line)

	}

	return status, err
}
