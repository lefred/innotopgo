package innotop

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/text"
)

func getMetadaLocks(mydb *sql.DB, thread_id string) ([][]string, error) {
	stmt := `WITH mdl_lock_summary AS (
            SELECT
            owner_thread_id,
            GROUP_CONCAT(
            DISTINCT
            CONCAT(
            LOCK_STATUS, ' ',
            lock_type, ' on ',
            IF(object_type='USER LEVEL LOCK', CONCAT(object_name, ' (user lock)'), CONCAT(OBJECT_SCHEMA, '.', OBJECT_NAME))
            )
            ORDER BY object_type ASC, LOCK_STATUS ASC, lock_type ASC
            SEPARATOR '\n') AS lock_summary
            FROM performance_schema.metadata_locks
            GROUP BY owner_thread_id
            )
            SELECT
            mdl_lock_summary.lock_summary
            FROM sys.processlist ps
            INNER JOIN mdl_lock_summary ON ps.thd_id=mdl_lock_summary.owner_thread_id
            WHERE conn_id`
	stmt = fmt.Sprintf("%s=%s;", stmt, thread_id)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, err
	}
	_, data, err := db.GetData(rows)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 {
		err = errors.New("not found")
		return nil, err
	}

	return data, err
}

func getDadaLocks(mydb *sql.DB, thread_id string) ([][]string, error) {
	stmt := `SELECT OBJECT_SCHEMA, OBJECT_NAME, LOCK_TYPE,
                         LOCK_MODE, LOCK_STATUS, INDEX_NAME, GROUP_CONCAT(LOCK_DATA SEPARATOR '|')
                         FROM INFORMATION_SCHEMA.INNODB_TRX
                         JOIN performance_schema.data_locks d
                           ON d.ENGINE_TRANSACTION_ID = trx_id
                         WHERE trx_mysql_thread_id`
	stmt = fmt.Sprintf("%s=%s GROUP BY 1,2,3,4,5,6 ORDER BY 1,2, 3 DESC, 6;", stmt, thread_id)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, err
	}
	_, data, err := db.GetData(rows)
	if err != nil {
		return nil, err
	}
	if len(data) < 1 {
		err = errors.New("not found")
		return nil, err
	}

	return data, err
}

func GetQueryConnByThreadId(mydb *sql.DB, thread_id string) (string, string, error) {
	stmt := fmt.Sprintf("select conn_id, current_statement from sys.x$processlist where thd_id=%s;", thread_id)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return "", "", err
	}
	_, data, err := db.GetData(rows)
	if err != nil {
		return "", "", err
	}
	if len(data) < 1 {
		err = errors.New("not found")
		return "", "", err
	}
	var query_text string
	var conn_id string
	for _, row := range data {
		conn_id = row[0]
		query_text = row[1]
	}

	return conn_id, query_text, err
}

func DisplayLocking(ctx context.Context, mydb *sql.DB, c *container.Container, top_window *text.Text, main_window *text.Text, thread_id string) error {
	conn_id, query_text, err := GetQueryConnByThreadId(mydb, thread_id)
	if err != nil {
		return err
	}
	data, err := getMetadaLocks(mydb, conn_id)
	if err != nil {
		return err
	}
	top_window.Reset()
	if len(query_text) > 0 {
		top_window.Write(query_text)
	} else {
		top_window.Write("no running query... (sleep)")
	}
	main_window.Reset()
	main_window.Write("\n")
	main_window.Write("Metadata Locks:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
	main_window.Write("\n\n")
	for _, row := range data {
		main_window.Write(row[0])
		main_window.Write("\n")
	}

	data, err = getDadaLocks(mydb, conn_id)
	if err != nil {
		return err
	}
	main_window.Write("\n")
	main_window.Write("Data Locks:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
	main_window.Write("\n\n")
	for _, row := range data {
		if len(row[5]) < 1 {
			main_window.Write(fmt.Sprintf("%v %v (%v) LOCK ON %v.%v", row[4], row[2], row[3], row[0], row[1]))
		} else {
			// we won't print it directly
			to_print := fmt.Sprintf("%v %v (%v) LOCK ON %v.%v [%v] ", row[4], row[2], row[3], row[0], row[1], row[5])
			str_len := len(to_print)
			main_window.Write(to_print)
			var cols []string
			var pk_cols []string
			stmt := `SELECT ifi.name, ifi.pos, ii.name, ifi.index_id
                                    FROM INFORMATION_SCHEMA.INNODB_TABLES it
                                    LEFT JOIN INFORMATION_SCHEMA.INNODB_INDEXES ii
                                           ON ii.table_id = it.table_id AND
                                              (ii.name = '`
			stmt = stmt + row[5]
			stmt = stmt + `' OR ii.name='PRIMARY')
                                   LEFT JOIN INFORMATION_SCHEMA.INNODB_FIELDS ifi
                                           ON ifi.index_id = ii.index_id
                                   WHERE it.name = '`
			stmt = stmt + row[0] + "/" + row[1]
			stmt = stmt + `' ORDER BY ii.NAME, POS`
			rows, err := db.Query(mydb, stmt)
			if err != nil {
				return err
			}
			_, data2, err := db.GetData(rows)
			if err != nil {
				return err
			}
			for _, row2 := range data2 {
				if row2[2] == "PRIMARY" {
					pk_cols = append(pk_cols, row2[0])
				} else {
					cols = append(cols, row2[0])
				}
			}
			// if there is an index name
			if len(row[5]) > 1 {
				records := strings.Split(row[6], "|")
				next_line := false
				for _, record := range records {
					if next_line {
						main_window.Write("\n")
						main_window.Write(strings.Repeat(" ", str_len))
					}
					columns := strings.Split(record, ", ")
					comma_str := ""
					column_to_disp := ""
					if row[5] != "PRIMARY" {
						i := 0
						main_window.Write(" (")
						for _, column := range columns {
							if len(column) == 10 && strings.HasPrefix(column, "0x") {
								val, _ := strconv.ParseUint(column[2:], 16, 32)
								column_to_disp = fmt.Sprintf("%v", val>>5)
							} else {
								column_to_disp = strings.TrimSpace(column)
							}
							if i < len(columns)-2 {
								comma_str = ", "
							} else {
								comma_str = ""
							}
							if i < len(cols) {
								main_window.Write(fmt.Sprintf("%v=%v%v", cols[i], column_to_disp, comma_str))
							}
							i++
						}
						if len(pk_cols) > 0 {
							main_window.Write(fmt.Sprintf(") => (%v=%v)", pk_cols[0], columns[i-1]))
						}
					} else {
						main_window.Write("(")
						i := 0
						for _, column := range columns {
							if len(column) == 10 && strings.HasPrefix(column, "0x") {
								val, _ := strconv.ParseUint(column[2:], 16, 32)
								column_to_disp = fmt.Sprintf("%v", val>>5)
							} else {
								column_to_disp = strings.TrimSpace(column)
							}
							if i < (len(columns) - 1) {
								comma_str = ", "
							} else {
								comma_str = ""
							}
							if i < len(pk_cols) {
								if len(columns) < len(pk_cols) {
									main_window.Write("<")
									j := 0
									for _, pk_el := range pk_cols {
										if j < (len(pk_cols) - 1) {
											comma_str = ", "
										} else {
											comma_str = ""
										}
										main_window.Write(pk_el + comma_str)
										j++
									}
									main_window.Write(">=")
									main_window.Write(column_to_disp)
								} else {
									main_window.Write(pk_cols[i] + "=" + column + comma_str)
								}
							}
							i++
						}
						main_window.Write(")")
					}
					next_line = true
				}

			}

		}
		main_window.Write("\n")
	}
	main_window.Write("\n\n")
	// Get Statement Blocking Us
	stmt := "SELECT REPLACE(locked_table,'`','') `TABLE`" +
		", locked_type, PROCESSLIST_INFO, waiting_lock_mode," +
		"          waiting_trx_rows_locked, waiting_trx_started, wait_age_secs, blocking_pid," +
		"          last_statement" +
		"   FROM performance_schema.threads AS t" +
		"   JOIN sys.innodb_lock_waits AS ilw ON ilw.waiting_pid = t.PROCESSLIST_ID " +
		"   JOIN sys.processlist proc ON proc.conn_id = blocking_pid" +
		"   WHERE waiting_pid=" + conn_id
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return err
	}
	_, data, err = db.GetData(rows)
	if err != nil {
		return err
	}
	for _, row := range data {
		main_window.Write("Last statement of the blocking trx:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
		main_window.Write("\n\n")
		main_window.Write(fmt.Sprintf("BLOCKED FOR %v SECONDS BY (mysql_thread_id: %v)\n\n", row[6], row[7]))
		main_window.Write(row[8], text.WriteCellOpts(cell.Italic()))
	}
	// Get Statements We Are Blocking
	stmt = "SELECT REPLACE(locked_table,'`','') `TABLE`, locked_type, waiting_query, waiting_lock_mode," +
		"             waiting_trx_rows_locked, waiting_trx_started, wait_age_secs, processlist_id" +
		"       FROM performance_schema.threads AS t" +
		"       JOIN sys.innodb_lock_waits AS ilw" +
		"         ON ilw.waiting_pid = t.PROCESSLIST_ID where blocking_pid=" + conn_id
	rows, err = db.Query(mydb, stmt)
	if err != nil {
		return err
	}
	_, data, err = db.GetData(rows)
	if err != nil {
		return err
	}
	for _, row := range data {
		main_window.Write("Statement we are blocking:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
		main_window.Write("\n\n")
		main_window.Write(fmt.Sprintf("BLOCKING %v (%v) LOCK ON %v FOR %v SECONDS (mysql_thread_id: %v)\n\n",
			row[1], row[3], row[0], row[6], row[7]))
		main_window.Write(row[2], text.WriteCellOpts(cell.Italic()))
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
	c.Update("main_container", container.BorderTitle("LOCKING INFORMATION (<-- <Backspace> to return)"))

	return nil
}
