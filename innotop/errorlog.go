package innotop

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/elliotchance/orderedmap"
	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/button"
	"github.com/mum4k/termdash/widgets/text"
)

func refresh_errorlog_info(t *tcell.Terminal, cancel context.CancelFunc, ctx context.Context, interval time.Duration, fn func() error) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				t.Close()
				ExitWithError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func GetErrorLog(mydb *sql.DB, choices_prio_info *orderedmap.OrderedMap, choices_sub_info *orderedmap.OrderedMap) ([]string, [][]string, error) {
	prio_string := ""
	for _, key := range choices_prio_info.Keys() {
		element, _ := choices_prio_info.Get(key)
		if element == false {
			prio_string = prio_string + ",'" + fmt.Sprintf("%v", key) + "'"
		}
	}
	prio_string_query := ""
	if len(prio_string) > 0 {
		prio_string_query = fmt.Sprintf("prio NOT IN (%v)", prio_string[1:])
	}

	sub_string := ""
	for _, key := range choices_sub_info.Keys() {
		element, _ := choices_sub_info.Get(key)
		if element == false {
			sub_string = sub_string + ",'" + fmt.Sprintf("%v", key) + "'"
		}
	}
	sub_string_query := ""
	if len(sub_string) > 0 {
		sub_string_query = fmt.Sprintf("subsystem NOT IN (%v)", sub_string[1:])
	}

	if len(prio_string_query) > 0 {
		prio_string_query = "WHERE " + prio_string_query
		if len(sub_string_query) > 0 {
			sub_string_query = "AND " + sub_string_query
		}
	} else {
		if len(sub_string_query) > 0 {
			prio_string_query = "WHERE "
		}
	}

	stmt := fmt.Sprintf("SELECT *, cast(unix_timestamp(logged)*1000000 as unsigned) logged_int FROM performance_schema.error_log %v %v ORDER BY logged", prio_string_query, sub_string_query)

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

func DisplayErrorlog(mydb *sql.DB, c *container.Container, t *tcell.Terminal) (keyboard.Key, error) {
	ctxmem, cancel := context.WithCancel(context.Background())
	k := keyboard.KeyBackspace2
	prio_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}
	subsystem_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	errorlog_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	choices_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}
	choices_prio_info := orderedmap.NewOrderedMap()
	choices_sub_info := orderedmap.NewOrderedMap()

	choices_prio_info.Set("system", true)
	choices_prio_info.Set("note", true)
	choices_prio_info.Set("warning", true)
	choices_prio_info.Set("error", true)

	choices_sub_info.Set("server", true)
	choices_sub_info.Set("innodb", true)
	choices_sub_info.Set("repl", true)

	reset_window := true
	last_logged := 0

	go refresh_errorlog_info(t, cancel, ctxmem, 1*time.Second, func() error {
		choices_window.Reset()
		for _, key := range choices_prio_info.Keys() {
			element_p := " "
			element, _ := choices_prio_info.Get(key)
			if element == true {
				element_p = "X"
			}
			choices_window.Write(fmt.Sprintf("%10v : [%1v]\n", key, element_p))
		}
		choices_window.Write("\n")
		for _, key := range choices_sub_info.Keys() {
			element_p := " "
			element, _ := choices_sub_info.Get(key)
			if element == true {
				element_p = "X"
			}
			choices_window.Write(fmt.Sprintf("%10v : [%1v]\n", key, element_p))
		}
		if reset_window {
			errorlog_window.Reset()
			last_logged = 0
			reset_window = false
		}

		_, data, err := GetErrorLog(mydb, choices_prio_info, choices_sub_info)
		if err != nil {
			cancel()
			return err
		}
		for _, row := range data {
			new_last_logged, _ := strconv.Atoi(row[6])
			if new_last_logged > last_logged {
				if row[2] == "Error" {
					errorlog_window.Write(fmt.Sprintf("%26v %5v %7v %9v %6v %v\n", row[0], row[1], row[2], row[3], row[4], row[5]), text.WriteCellOpts(cell.FgColor(cell.ColorRed)))
				} else {
					if row[2] == "Warning" {
						errorlog_window.Write(fmt.Sprintf("%26v %5v %7v %9v %6v %v\n", row[0], row[1], row[2], row[3], row[4], row[5]), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(172))))
					} else {
						errorlog_window.Write(fmt.Sprintf("%26v %5v %7v %9v %6v %v\n", row[0], row[1], row[2], row[3], row[4], row[5]))
					}
				}
				last_logged = new_last_logged
			}
		}
		return nil
	})

	systemB, err := button.New("(s)ystem", func() error {
		value, _ := choices_prio_info.Get("system")
		if value == true {
			choices_prio_info.Set("system", false)
		} else {
			choices_prio_info.Set("system", true)
		}
		reset_window = true
		return nil
	},
		button.GlobalKey('s'),
		button.WidthFor("(w)arning"),
	)
	if err != nil {
		cancel()
		return k, err
	}
	noteB, err := button.New("(n)note", func() error {
		value, _ := choices_prio_info.Get("note")
		if value == true {
			choices_prio_info.Set("note", false)
		} else {
			choices_prio_info.Set("note", true)
		}
		reset_window = true
		return nil
	},
		button.GlobalKey('n'),
		button.WidthFor("(w)arning"),
	)
	if err != nil {
		cancel()
		return k, err
	}
	warningB, err := button.New("(w)warning", func() error {
		value, _ := choices_prio_info.Get("warning")
		if value == true {
			choices_prio_info.Set("warning", false)
		} else {
			choices_prio_info.Set("warning", true)
		}
		reset_window = true
		return nil
	},
		button.GlobalKey('w'),
	)
	if err != nil {
		cancel()
		return k, err
	}
	errorB, err := button.New("e(r)ror", func() error {
		value, _ := choices_prio_info.Get("error")
		if value == true {
			choices_prio_info.Set("error", false)
		} else {
			choices_prio_info.Set("error", true)
		}
		reset_window = true
		return nil
	},
		button.GlobalKey('r'),
		button.WidthFor("(w)arning"),
	)
	if err != nil {
		cancel()
		return k, err
	}

	serverB, err := button.New("server", func() error {
		value, _ := choices_sub_info.Get("server")
		if value == true {
			choices_sub_info.Set("server", false)
		} else {
			choices_sub_info.Set("server", true)
		}
		reset_window = true
		return nil
	},
		button.WidthFor("replication"),
		button.FillColor(cell.ColorNumber(15)),
	)
	if err != nil {
		cancel()
		return k, err
	}
	innodbB, err := button.New("InnoDB", func() error {
		value, _ := choices_sub_info.Get("innodb")
		if value == true {
			choices_sub_info.Set("innodb", false)
		} else {
			choices_sub_info.Set("innodb", true)
		}
		reset_window = true
		return nil
	},
		button.WidthFor("replication"),
		button.FillColor(cell.ColorNumber(15)),
	)
	if err != nil {
		cancel()
		return k, err
	}
	replB, err := button.New("replication", func() error {
		value, _ := choices_sub_info.Get("repl")
		if value == true {
			choices_sub_info.Set("repl", false)
		} else {
			choices_sub_info.Set("repl", true)
		}
		reset_window = true
		return nil
	},
		button.FillColor(cell.ColorNumber(15)),
	)
	if err != nil {
		cancel()
		return k, err
	}

	c.Update("dyn_top_container",
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.ID("top_container"),
								container.PlaceWidget(prio_window),
								container.FocusedColor(cell.ColorNumber(15)),
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.PlaceWidget(systemB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
											container.Right(
												container.PlaceWidget(noteB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
										),
									),
									container.Right(
										container.SplitVertical(
											container.Left(
												container.PlaceWidget(warningB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
											container.Right(
												container.PlaceWidget(errorB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
										),
									),
								),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.ID("top2_container"),
								container.PlaceWidget(subsystem_window),
								container.FocusedColor(cell.ColorNumber(15)),
								container.SplitVertical(
									container.Left(
										container.SplitVertical(
											container.Left(
												container.PlaceWidget(serverB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
											container.Right(
												container.PlaceWidget(innodbB),
												container.AlignHorizontal(align.HorizontalCenter),
											),
										),
									),
									container.Right(
										container.PlaceWidget(replB),
										container.AlignHorizontal(align.HorizontalCenter),
									),
									container.SplitPercent(75),
								),
							),
							container.SplitPercent(50),
						),
					),
					container.Right(
						container.Border(linestyle.Light),
						container.ID("choices_container"),
						container.PlaceWidget(choices_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.SplitPercent(80),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.ID("error_container"),
				container.PlaceWidget(errorlog_window),
				container.FocusedColor(cell.ColorNumber(15)),
			),
			container.SplitFixed(12),
		),
	)
	c.Update("error_container", container.Focused())
	c.Update("top_container", container.BorderTitle("Error Log - Priorities (<-- <Backspace> to return to Processlist)"))
	c.Update("top2_container", container.BorderTitle("Subsystems"))
	c.Update("error_container", container.BorderTitle("Error Events"))
	//tot_mem_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

	quitter := func(k2 *terminalapi.Keyboard) {
		if k2.Key == keyboard.KeyEsc || k2.Key == keyboard.KeyCtrlC {
			k = k2.Key
			cancel()
			return
		} else if k2.Key == keyboard.KeyBackspace2 {
			k = k2.Key
			cancel()
			return
		} else {
			return
		}
	}

	if err := termdash.Run(ctxmem, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		cancel()
		t.Close()
		return k, err
	}
	return k, nil
}
