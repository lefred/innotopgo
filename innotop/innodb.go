package innotop

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/text"
)

func refresh_innodb_info(t *tcell.Terminal, cancel context.CancelFunc, ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				cancel()
				t.Close()
				ExitWithError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func DisplayInnoDB(mydb *sql.DB, c *container.Container, t *tcell.Terminal) (keyboard.Key, error) {
	ctx, cancel := context.WithCancel(context.Background())
	k := keyboard.KeyBackspace2
	details_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}
	info_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	bp_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(31))),
		donut.Label("Buffer Pool %", cell.FgColor(cell.ColorNumber(31))),
	)
	if err != nil {
		cancel()
		return k, err
	}

	bp_read_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorPurple)),
		donut.Label("Disk Read Ratio %", cell.FgColor(cell.ColorPurple)),
	)
	if err != nil {
		cancel()
		return k, err
	}

	redo_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(172))),
		donut.Label("Checkpoint Age %", cell.FgColor(cell.ColorNumber(172))),
	)
	if err != nil {
		cancel()
		return k, err
	}

	ahi_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("AHI Ratio %", cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		cancel()
		return k, err
	}

	top_window, err := text.New(text.WrapAtWords())
	if err != nil {
		cancel()
		return k, err
	}

	var prev_innodb_status = make(map[string]string)

	go refresh_innodb_info(t, cancel, ctx, 1*time.Second, func() error {
		cols, data, err := GetBPFill(mydb)
		if err != nil {
			cancel()
			return err
		}
		var bp_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				bp_info[cols[i]] = row[i]
			}
		}
		cols, data, err = GetRedoCapacity(mydb)
		if err != nil {
			cancel()
			return err
		}
		var redo_capacity = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				redo_capacity[cols[i]] = row[i]
			}
		}
		var innodb_redo_log_capacity = 0
		var use_innodb_redo_log_capacity = false
		if redo_capacity["InnoDBLogCapacity"] != "" {
			innodb_redo_log_capacity, _ = strconv.Atoi(redo_capacity["InnoDBLogCapacityRaw"])
			use_innodb_redo_log_capacity = true
		} else {
			innodb_redo_log_file_size_raw, _ := strconv.Atoi(redo_capacity["InnoDBLogFileSizeRaw"])
			innodb_redo_log_files, _ := strconv.Atoi(redo_capacity["NbFiles"])
			innodb_redo_log_capacity = innodb_redo_log_file_size_raw * innodb_redo_log_files
		}

		cols, data, err = GetRedoInfo(mydb, innodb_redo_log_capacity)
		if err != nil {
			cancel()
			return err
		}
		var redo_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				redo_info[cols[i]] = row[i]
			}
		}
		cols, data, err = GetAHI(mydb)
		if err != nil {
			cancel()
			return err
		}
		var ahi_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				ahi_info[cols[i]] = row[i]
			}
		}
		_, data, err = GetInnoDBStatus(mydb)
		if err != nil {
			cancel()
			return err
		}
		var innodb_status = make(map[string]string)
		for _, row := range data {
			innodb_status[row[0]] = row[1]
		}

		graph_pct, _ := strconv.Atoi(bp_info["BufferPoolFull"])
		bp_graph.Percent(graph_pct)
		chkpt_pct, _ := strconv.Atoi(redo_info["CheckpointAgeInt"])
		redo_graph.Percent(chkpt_pct)
		graph_pct, _ = strconv.Atoi(ahi_info["AHIRatioInt"])
		ahi_graph.Percent(graph_pct)
		graph_pct, _ = strconv.Atoi(bp_info["DiskReadRatioInt"])
		bp_read_graph.Percent(graph_pct)

		uptime_sec, _ := strconv.Atoi(redo_info["Uptime"])
		top_window.Reset()
		top_window.Write("\n")
		top_window.Write(PrintLabel("Buffer Pool Size"))
		top_window.Write(fmt.Sprintf("%-10v", bp_info["BP_Size"]))
		top_window.Write(PrintLabel("Uptime"))
		top_window.Write(fmt.Sprintf("%-10v", (time.Duration(uptime_sec) * time.Second)))
		top_window.Write("\n")
		top_window.Write(PrintLabel("Buffer Pool Instances"))
		top_window.Write(fmt.Sprintf("%-10v", bp_info["BP_instances"]))
		top_window.Write("\n\n")
		top_window.Write(PrintLabel("Redo Log"))
		top_window.Write(fmt.Sprintf("%-10v", redo_info["RedoEnabled"]))
		top_window.Write("\n")
		if redo_info["RedoEnabled"] == "ON" {
			if use_innodb_redo_log_capacity == true {
				top_window.Write(PrintLabel("InnodDB Log Capacity"))
				top_window.Write(fmt.Sprintf("%-10v", redo_capacity["InnoDBLogCapacity"]))
			} else {
				top_window.Write(PrintLabel("InnodDB Log File Size"))
				top_window.Write(fmt.Sprintf("%-10v", redo_capacity["InnoDBLogFileSize"]))
				top_window.Write("\n")
				top_window.Write(PrintLabel("Num InnoDB Log File"))
				top_window.Write(fmt.Sprintf("%-10v", redo_capacity["NbFiles"]))
			}
			top_window.Write("\n")
			top_window.Write(PrintLabel("Checkpoint Info"))
			top_window.Write(fmt.Sprintf("%-25v", redo_info["CheckpointInfo"]))
			top_window.Write("\n")
			top_window.Write(PrintLabel("Checkpoint Age"))
			color := cell.ColorDefault
			if chkpt_pct > 80 {
				color = cell.ColorRed
			} else if chkpt_pct > 70 {
				color = cell.ColorNumber(172)
			}
			top_window.Write(redo_info["CheckpointAge"]+"%", text.WriteCellOpts(cell.FgColor(color)))
		}
		top_window.Write("\n\n")
		top_window.Write(PrintLabel("Adaptive Hash Index"))
		top_window.Write(fmt.Sprintf("%-10v", ahi_info["AHIEnabled"]))
		top_window.Write("\n")
		if ahi_info["AHIEnabled"] == "ON" {
			top_window.Write(PrintLabel("Num AHI Partitions"))
			top_window.Write(fmt.Sprintf("%-10v", ahi_info["AHIParts"]))
		}

		// Display information in the bottom frame
		info_window.Reset()
		info_window.Write(fmt.Sprintf("%v\n", innodb_status["Innodb_buffer_pool_dump_status"]))
		info_window.Write(fmt.Sprintf("%v\n", innodb_status["Innodb_buffer_pool_load_status"]))
		info_window.Write(fmt.Sprintf("%v\n", innodb_status["Innodb_buffer_pool_resize_status"]))

		// Display status in the details window per second
		// Calculation is required and compare between previous run
		details_window.Reset()
		details_window.Write("\n")
		details_window.Write(PrintLabel("Read Requests"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_buffer_pool_read_requests")))
		details_window.Write(PrintLabel("Disk Reads"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_buffer_pool_reads")))
		details_window.Write("\n")
		details_window.Write(PrintLabel("Write Requests"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_buffer_pool_write_requests")))
		details_window.Write("\n\n")
		details_window.Write(PrintLabel("Dirty Data"))
		dirty, _ := strconv.Atoi(innodb_status["Innodb_buffer_pool_bytes_dirty"])
		details_window.Write(fmt.Sprintf("%10v", FormatBytes(dirty)))
		details_window.Write(PrintLabel("Dirty Data/sec"))
		details_window.Write(fmt.Sprintf("%10v", FormatBytes(
			GetValue(prev_innodb_status, innodb_status, "Innodb_buffer_pool_bytes_dirty"))))
		details_window.Write("\n\n")
		details_window.Write(PrintLabel("Pending Reads"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_data_pending_reads")))
		details_window.Write(PrintLabel("Pending Fsync"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_data_pending_fsyncs")))
		details_window.Write("\n")
		details_window.Write(PrintLabel("Pending Writes"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_data_pending_writes")))
		details_window.Write("\n\n")
		details_window.Write(PrintLabel("OS Log Pending Writes"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_os_log_pending_writes")))
		details_window.Write(PrintLabel("OS Log Pending Fsyncs"))
		details_window.Write(fmt.Sprintf("%10v",
			GetValue(prev_innodb_status, innodb_status, "Innodb_os_log_pending_fsyncs")))

		prev_innodb_status = innodb_status
		return nil
	})

	c.Update("dyn_top_container",
		container.SplitVertical(
			container.Left(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.ID("top_container"),
						container.PlaceWidget(top_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.Bottom(
						container.SplitHorizontal(
							container.Top(
								container.Border(linestyle.Light),
								container.ID("main_container"),
								container.PlaceWidget(details_window),
								container.FocusedColor(cell.ColorNumber(15)),
							),
							container.Bottom(
								container.Border(linestyle.Light),
								container.ID("bottom_container"),
								container.PlaceWidget(info_window),
								container.FocusedColor(cell.ColorNumber(15)),
							),
							container.SplitPercent(70),
						),
					),
					container.SplitFixed(15),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.ID("left_graph1"),
								container.FocusedColor(cell.ColorNumber(15)),
								container.PlaceWidget(redo_graph),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.ID("right_graph1"),
								container.FocusedColor(cell.ColorNumber(15)),
								container.PlaceWidget(bp_graph),
							),
							container.SplitPercent(50),
						),
					),
					container.Bottom(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.ID("left_graph2"),
								container.FocusedColor(cell.ColorNumber(15)),
								container.PlaceWidget(ahi_graph),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.ID("right_graph2"),
								container.FocusedColor(cell.ColorNumber(15)),
								container.PlaceWidget(bp_read_graph),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(50),
				),
			),
			container.SplitFixed(85),
		),
	)
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("top_container", container.BorderTitle("InnoDB Info (<-- <Backspace> to return to Processlist)"))
	c.Update("main_container", container.BorderTitle("InnoDB Buffer Pool"))
	top_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

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
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		cancel()
		t.Close()
		return k, err
	}
	return k, nil
}
