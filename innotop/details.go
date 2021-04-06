package innotop

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/widgets/text"
)

func GetDetailsByThreadId(mydb *sql.DB, thread_id string) ([]string, [][]string, error) {
	stmt := fmt.Sprintf(`select
							pps.THREAD_ID, TYPE, pps.PROCESSLIST_ID, pps.PROCESSLIST_COMMAND,
							pps.PROCESSLIST_STATE, pps.PARENT_THREAD_ID,
							pps.INSTRUMENTED, pps.HISTORY,
							pps.CONNECTION_TYPE, THREAD_OS_ID,
							RESOURCE_GROUP, ISOLATION_LEVEL, AUTOCOMMIT, user,
							conattr_progname.ATTR_NAME, conattr_progname.ATTR_VALUE,
							format_bytes(current_allocated) AS curr_mem_alloc,
							format_bytes(current_avg_alloc) AS curr_avg_mem_alloc,
							format_bytes(current_max_alloc) AS curr_max_mem_alloc,
							format_bytes(total_allocated) AS total_mem_allocated,
							SQL_TEXT,
							ERRORS, WARNINGS, ROWS_AFFECTED, ROWS_SENT, ROWS_EXAMINED,
							CREATED_TMP_DISK_TABLES, CREATED_TMP_TABLES, SELECT_FULL_JOIN,
							SELECT_RANGE, SELECT_RANGE_CHECK, SELECT_SCAN,
							SORT_MERGE_PASSES, SORT_RANGE, SORT_ROWS, SORT_SCAN,
							NO_INDEX_USED, NO_GOOD_INDEX_USED
								 from (((((((performance_schema.threads pps
		                          left join performance_schema.events_waits_current ewc
								    on((pps.THREAD_ID = ewc.THREAD_ID)))
								  left join performance_schema.events_stages_current estc
								    on((pps.THREAD_ID = estc.THREAD_ID)))
								  left join performance_schema.events_statements_current esc
								    on((pps.THREAD_ID = esc.THREAD_ID)))
								  left join performance_schema.events_transactions_current etc
								    on((pps.THREAD_ID = etc.THREAD_ID)))
								  left join sys.x$memory_by_thread_by_current_bytes mem
								    on((pps.THREAD_ID = mem.thread_id)))
								  left join performance_schema.session_connect_attrs conattr_pid
								    on(((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID)
									   and (conattr_pid.ATTR_NAME = '_pid'))))
								  left join performance_schema.session_connect_attrs conattr_progname
								    on(((conattr_progname.PROCESSLIST_ID = pps.PROCESSLIST_ID)
									   and (conattr_progname.ATTR_NAME = 'program_name'))))
								  where pps.PROCESSLIST_ID is not null
								       and pps.PROCESSLIST_COMMAND <> 'Daemon'
									   and pps.THREAD_ID=%s;`, thread_id)
	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}
	return cols, data, err
}

func PrintLabel(label string, col_opt ...int) (string, text.WriteOption) {
	col := 0
	if len(col_opt) > 0 {
		col = col_opt[0]
	}
	tot_col := col * 27
	if tot_col > 0 {
		tot_col = tot_col + 15
	}
	out_col := strings.Repeat(" ", tot_col)

	out_label := fmt.Sprintf("%s%27s: ", out_col, label)
	out_opts := text.WriteCellOpts(cell.Bold())
	return out_label, out_opts
}

func DisplayThreadDetails(mydb *sql.DB, c *container.Container, thread_id string) error {
	details_window, err := text.New()
	if err != nil {
		return err
	}
	c.Update("dyn_top_container", container.SplitHorizontal(container.Top(
		container.Border(linestyle.Light),
		container.ID("top_container"),
	),
		container.Bottom(
			container.Border(linestyle.Light),
			container.ID("main_container"),
			container.PlaceWidget(details_window),
			container.FocusedColor(cell.ColorNumber(15)),
		), container.SplitFixed(0)))
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("Thread Details (<-- <Backspace> to return to Processlist)"))
	details_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
	cols, data, err := GetDetailsByThreadId(mydb, thread_id)
	if err != nil {
		return err
	}
	if len(data) < 1 {
		err = errors.New("not found")
		return err
	}
	var details = make(map[string]string)
	details_window.Reset()
	for _, row := range data {
		for i := 0; i < len(row); i++ {
			name_value := cols[i]
			for {
				// this is a loop to change the name of a column
				// when they are multiple with the same name in the result set.
				// example: PROCESSLIST_ID, it there is another one, the second
				//          will be called PROCESSLIST_ID_ and the third one
				//          PROCESSLIST_ID__
				if _, ok := details[name_value]; ok {
					name_value = name_value + "_"
				} else {
					break
				}
			}
			details[name_value] = row[i]
		}

	}

	details_window.Write("\n")
	details_window.Write(PrintLabel("Thread_id"))
	details_window.Write(fmt.Sprintf("%-15v", details["THREAD_ID"]))
	details_window.Write(PrintLabel("Pid"))
	details_window.Write(fmt.Sprintf("%-15v", details["PROCESSLIST_ID"]))
	details_window.Write(PrintLabel("Os thread_id"))
	details_window.Write(fmt.Sprintf("%-5v\n", details["THREAD_OS_ID"]))

	details_window.Write(PrintLabel("Type"))
	details_window.Write(fmt.Sprintf("%-15v", details["TYPE"]))
	details_window.Write(PrintLabel("Command"))
	details_window.Write(fmt.Sprintf("%-15v", details["PROCESSLIST_COMMAND"]))
	details_window.Write(PrintLabel("State"))
	details_window.Write(fmt.Sprintf("%-10v\n\n", details["PROCESSLIST_STATE"]))

	details_window.Write(PrintLabel("Connection Type"))
	details_window.Write(fmt.Sprintf("%-15v", details["CONNECTION_TYPE"]))
	details_window.Write(PrintLabel("Program Name"))
	details_window.Write(fmt.Sprintf("%-15v", details["ATTR_VALUE"]))
	details_window.Write(PrintLabel("User"))
	details_window.Write(fmt.Sprintf("%-30v\n\n", details["user"]))

	details_window.Write(PrintLabel("Instrumented"))
	details_window.Write(fmt.Sprintf("%-15v", details["INSTRUMENTED"]))
	details_window.Write(PrintLabel("Isolation Level"))
	details_window.Write(fmt.Sprintf("%-15v", details["ISOLATION_LEVEL"]))
	details_window.Write(PrintLabel("Resource Group"))
	details_window.Write(fmt.Sprintf("%-30v\n", details["RESOURCE_GROUP"]))

	details_window.Write(PrintLabel("History"))
	details_window.Write(fmt.Sprintf("%-15v", details["HISTORY"]))
	details_window.Write(PrintLabel("Autocommit"))
	details_window.Write(fmt.Sprintf("%-20v\n\n", details["AUTOCOMMIT"]))

	details_window.Write(PrintLabel("Current Mem Alloc"))
	details_window.Write(fmt.Sprintf("%15v", details["curr_mem_alloc"]))
	details_window.Write(PrintLabel("Row Affected", 1))
	details_window.Write(fmt.Sprintf("%15v", details["ROWS_AFFECTED"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("Current Avg Mem Alloc"))
	details_window.Write(fmt.Sprintf("%15v", details["curr_avg_mem_alloc"]))
	details_window.Write(PrintLabel("Rows Examined", 1))
	details_window.Write(fmt.Sprintf("%15v", details["ROWS_EXAMINED"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("Current Max Mem Alloc"))
	details_window.Write(fmt.Sprintf("%15v", details["curr_max_mem_alloc"]))
	details_window.Write(PrintLabel("Rows Sent", 1))
	details_window.Write(fmt.Sprintf("%15v", details["ROWS_SENT"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("Total Mem Allocated"))
	details_window.Write(fmt.Sprintf("%15v", details["total_mem_allocated"]))
	details_window.Write("\n\n")
	details_window.Write(PrintLabel("Warnings"))
	details_window.Write(fmt.Sprintf("%15v", details["WARNINGS"]))
	details_window.Write(PrintLabel("Created Temp Tables", 1))
	details_window.Write(fmt.Sprintf("%7v", details["CREATED_TMP_TABLES"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("Errors"))
	details_window.Write(fmt.Sprintf("%15v", details["ERRORS"]))
	details_window.Write(PrintLabel("Created Temp Tables To Disk", 1))
	details_window.Write(fmt.Sprintf("%7v", details["CREATED_TMP_DISK_TABLES"]))
	details_window.Write("\n\n")
	details_window.Write(PrintLabel("Select Scan", 1))
	details_window.Write(fmt.Sprintf("%7v", details["SELECT_SCAN"]))
	details_window.Write(PrintLabel("Sort Scan"))
	details_window.Write(fmt.Sprintf("%7v", details["SORT_SCAN"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("No Index Used"))
	details_window.Write(fmt.Sprintf("%13v", details["NO_INDEX_USED"]))
	details_window.Write(PrintLabel("Select Full Join"))
	details_window.Write(fmt.Sprintf("%7v", details["SELECT_FULL_JOIN"]))
	details_window.Write(PrintLabel("Sort Rows"))
	details_window.Write(fmt.Sprintf("%7v", details["SORT_ROWS"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("No Good Index Used"))
	details_window.Write(fmt.Sprintf("%13v", details["NO_GOOD_INDEX_USED"]))
	details_window.Write(PrintLabel("Select Range"))
	details_window.Write(fmt.Sprintf("%7v", details["SELECT_RANGE"]))
	details_window.Write(PrintLabel("Sort Range"))
	details_window.Write(fmt.Sprintf("%7v", details["SORT_RANGE"]))
	details_window.Write("\n")
	details_window.Write(PrintLabel("Select Range Check", 1))
	details_window.Write(fmt.Sprintf("%7v", details["SELECT_RANGE_CHECK"]))
	details_window.Write("\n\n")
	details_window.Write(PrintLabel("Last Query"))
	details_window.Write(fmt.Sprintf("%v", details["SQL_TEXT"]), text.WriteCellOpts(cell.Italic()))

	return nil
}
