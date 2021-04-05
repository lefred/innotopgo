package innotop

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

func GetStatus(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `select variable_name, variable_value from performance_schema.global_status
	         union
			 select event_name, count_star
			 from performance_schema.events_statements_summary_global_by_event_name`
	rows := db.Query(mydb, stmt)
	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}

func GetComStmt(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SHOW GLOBAL STATUS LIKE 'Com_%'`
	rows := db.Query(mydb, stmt)
	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}

func DisplayStatus(mydb *sql.DB, top_window *text.Text, tlg *barchart.BarChart,
	trg *sparkline.SparkLine, prev_status map[string]string, old_values []int) (map[string]string, []int, error) {
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
	_, data, err = GetComStmt(mydb)
	if err != nil {
		panic(err)
	}
	var comstmt = make(map[string]string)
	for _, row := range data {
		comstmt[row[0]] = row[1]
	}

	uptime_sec, _ := strconv.Atoi(status["Uptime"])
	queries, _ := strconv.Atoi(status["Queries"])
	var values []int
	if prev_status != nil {
		prev_uptime_sec, _ := strconv.Atoi(prev_status["Uptime"])
		prev_queries, _ := strconv.Atoi(prev_status["Queries"])
		com_select, _ := strconv.Atoi(comstmt["Com_select"])
		com_insert, _ := strconv.Atoi(comstmt["Com_insert"])
		com_update, _ := strconv.Atoi(comstmt["Com_update"])
		com_delete, _ := strconv.Atoi(comstmt["Com_delete"])
		prev_com_select, _ := strconv.Atoi(prev_status["Com_select"])
		prev_com_insert, _ := strconv.Atoi(prev_status["Com_insert"])
		prev_com_update, _ := strconv.Atoi(prev_status["Com_update"])
		prev_com_delete, _ := strconv.Atoi(prev_status["Com_delete"])
		max_value := 10

		if (uptime_sec - prev_uptime_sec) < 1 {
			real_qps = 0
		} else {
			real_qps = (queries - prev_queries) / (uptime_sec - prev_uptime_sec)
			values = append(values, (com_select - prev_com_select))
			values = append(values, (com_insert - prev_com_insert))
			values = append(values, (com_update - prev_com_update))
			values = append(values, (com_delete - prev_com_delete))
			for i := 0; i < 4; i++ {
				if max_value < values[i] {
					max_value = values[i]
				}
				if len(old_values) > 0 {
					if max_value < old_values[i] {
						max_value = old_values[i]
					}
				}
			}
		}
		line = fmt.Sprintf(" Uptime: %-10v", (time.Duration(uptime_sec) * time.Second))
		top_window.Reset()
		top_window.Write(line)
		top_window.Write("\n")
		line = fmt.Sprintf("Threads: %v/%v (run/con)", status["Threads_running"], status["Threads_connected"])
		top_window.Write(line)
		top_window.Write("\n")
		line = fmt.Sprintf("    QPS: %-10v", (queries / uptime_sec))
		top_window.Write(line)
		line = fmt.Sprintf("real QPS: %v", real_qps)
		top_window.Write(line)
		trg.Add(([]int{real_qps}))
		tlg.Values(values, max_value)
	} else {
		line = fmt.Sprintf(" Uptime: %-10v", (time.Duration(uptime_sec) * time.Second))
		top_window.Reset()
		top_window.Write(line)
		top_window.Write("\n")
		line = fmt.Sprintf("Threads: %v/%v (run/con)", status["Threads_running"], status["Threads_connected"])
		top_window.Write(line)
		top_window.Write("\n")
		line = fmt.Sprintf("    QPS: %-10v", (queries / uptime_sec))
		top_window.Write(line)
	}
	combined_status := map[string]string{}
	for k, v := range status {
		combined_status[k] = v
	}
	for k, v := range comstmt {
		combined_status[k] = v
	}
	return combined_status, values, err
}
