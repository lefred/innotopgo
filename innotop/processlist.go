package innotop

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
)

const redrawInterval = 1000 * time.Millisecond

func Processlist(mydb *sql.DB, displaytype string) error {

	if displaytype == "simple" {
		cols, data, err := GetProcesslist(mydb)
		if err != nil {
			panic(err)
		}
		DisplaySimple(cols, data)
	} else {
		DisplayProcesslist(mydb)
	}
	return nil
}

func GetProcesslist(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `select pps.PROCESSLIST_COMMAND AS command,
                                  pps.THREAD_ID AS thd_id, pps.PROCESSLIST_ID AS conn_id,
                                  conattr_pid.ATTR_VALUE AS pid, pps.PROCESSLIST_STATE AS state,
                                  if((pps.NAME in ('thread/sql/one_connection','thread/thread_pool/tp_one_connection')),
                                   concat(pps.PROCESSLIST_USER,'@',pps.PROCESSLIST_HOST),
                                   replace(pps.NAME,'thread/','')) AS user,
                                  pps.PROCESSLIST_DB AS db, sys.format_statement(pps.PROCESSLIST_INFO) AS current_statement,
                                  if(isnull(esc.END_EVENT_ID), format_pico_time(esc.TIMER_WAIT),NULL) AS statement_latency,
                                  format_pico_time(esc.LOCK_TIME) AS lock_latency,
                                  if(isnull(esc.END_EVENT_ID),esc.TIMER_WAIT,0) AS sort_time
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
                                on(((conattr_pid.PROCESSLIST_ID = pps.PROCESSLIST_ID) and (conattr_pid.ATTR_NAME = '_pid'))))
                            left join performance_schema.session_connect_attrs conattr_progname
                                on(((conattr_progname.PROCESSLIST_ID = pps.PROCESSLIST_ID)
                                and (conattr_progname.ATTR_NAME = 'program_name'))))
                            where pps.PROCESSLIST_ID is not null
                              and pps.PROCESSLIST_COMMAND <> 'Daemon'
                              and user <> 'sql/event_scheduler'
                            order by sort_time desc
                        `
	rows := db.Query(mydb, stmt)
	cols, data, err := db.GetData(rows)
	if err != nil {
		panic(err)
	}

	return cols, data, err
}

func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func DisplayProcesslist(mydb *sql.DB) {

	show_processlist := true
	current_mode := "processlist"
	thread_id := "0"

	var c *container.Container

	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	innotop, err := text.New()
	if err != nil {
		panic(err)
	}
	main_window, err := text.New()
	if err != nil {
		panic(err)
	}

	bottom_input, err := textinput.New(
		textinput.MaxWidthCells(4),
		textinput.Label("Thread Id: ", cell.FgColor(cell.ColorNumber(31))),
		textinput.ClearOnSubmit(),
		textinput.OnSubmit(func(text string) error {
			// TODO: check if thread id is a number
			main_window.Reset()
			show_processlist = false
			thread_id = text
			err := DisplayExplain(mydb, c, main_window, text, "NORMAL")
			if err != nil {
				panic(err)
			}
			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	_, data, err := db.GetServerInfo(mydb)
	if err != nil {
		panic(err)
	}
	innotop.Write("Inno", text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7)), cell.FgColor(cell.ColorNumber(31)), cell.Bold()))
	innotop.Write("Top", text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7)), cell.FgColor(cell.ColorNumber(172)), cell.Bold()))
	innotop.Write(" Go | ", text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7)), cell.FgColor(cell.ColorNumber(31)), cell.Bold()))
	for _, row := range data {
		line := fmt.Sprintf("%s %s ", row[0], row[1])
		innotop.Write(line, text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7)), cell.FgColor(cell.ColorNumber(172)), cell.Italic()))
		line = fmt.Sprintf("[%s:%s]", row[2], row[3])
		innotop.Write(line, text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7)), cell.FgColor(cell.ColorNumber(31)), cell.Italic()))
	}
	innotop.Write(strings.Repeat(" ", 200), text.WriteCellOpts(cell.BgColor(cell.ColorNumber(7))))
	main_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
	go periodic(ctx, 1*time.Second, func() error {
		if show_processlist {

			_, data, err := GetProcesslist(mydb)
			if err != nil {
				panic(err)
			}
			main_window.Reset()
			header := fmt.Sprintf("%-7v %-5v %-5v %-7v %-15v %-20v %-12v %-10v %-10v %-65v\n", "Cmd", "Thd", "Conn", "Pid", "State", "User", "Db", "Time", "Lock Time", "Query")
			if err := main_window.Write(header, text.WriteCellOpts(cell.Bold())); err != nil {
				panic(err)
			}
			var color int
			for _, row := range data {
				line := fmt.Sprintf("%-7v %-5v %-5v %-7v %-15v %-20v %-12v %10v %10v %-65v\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[8], row[9], row[7])
				col_value, _ := strconv.Atoi(row[10])
				switch {
				case col_value > 60_000_000_000_000:
					color = 9 // red after 1min
				case col_value > 30_000_000_000_000:
					color = 172 // orange after 30sec
				case col_value > 10_000_000_000_000:
					color = 2 // green after 10sec
				case col_value > 5_000_000_000_000:
					color = 6 // blue after 5sec
				default:
					color = 15 // white
				}
				main_window.Write(line, text.WriteCellOpts(cell.FgColor(cell.ColorNumber(color))))
			}
		}
		return nil
	})

	c, err = container.New(
		t,
		container.SplitHorizontal(
			container.Top(
				container.PlaceWidget(innotop),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.ID("main_container"),
						container.BorderTitle("Processlist (ESC to quit)"),
						container.PlaceWidget(main_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.Bottom(
						container.ID("bottom_container"),
						container.AlignHorizontal(align.HorizontalLeft),
						container.Clear(),
					),
					container.SplitPercent(99),
				),
			),
			container.SplitFixed(1),
		),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == keyboard.KeyCtrlC {
			if show_processlist {
				cancel()
			}
		} else if k.Key == 'e' || k.Key == 'E' {
			c.Update("bottom_container", container.PlaceWidget(bottom_input))
			c.Update("bottom_container", container.Focused())
			current_mode = "explain_normal"
		} else if k.Key == keyboard.KeyBackspace2 {
			if !show_processlist {
				show_processlist = true
				c.Update("main_container", container.Focused())
				c.Update("main_container", container.BorderTitle("Processlist (ESC to quit)"))
				main_window.Reset()
				main_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
				current_mode = "processlist"
				thread_id = "0"
			}
		} else if k.Key == keyboard.KeySpace {
			if current_mode == "explain_normal" {
				main_window.Reset()
				err := DisplayExplain(mydb, c, main_window, thread_id, "FORMAT=TREE")
				if err != nil {
					panic(err)
				}
				current_mode = "explain_tree"
			} else if current_mode == "explain_tree" {
				main_window.Reset()
				err := DisplayExplain(mydb, c, main_window, thread_id, "FORMAT=JSON")
				if err != nil {
					panic(err)
				}
				current_mode = "explain_json"
			} else if current_mode == "explain_json" {
				main_window.Reset()
				err := DisplayExplain(mydb, c, main_window, thread_id, "NORMAL")
				if err != nil {
					panic(err)
				}
				current_mode = "explain_normal"
			}
		}

	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}
