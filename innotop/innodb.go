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

func refresh_innodb_info(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
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
	bp_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(31))),
		donut.Label("Buffer Pool %", cell.FgColor(cell.ColorNumber(31))),
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

	go refresh_innodb_info(ctx, 1*time.Second, func() error {
		cols, data, err := GetBPFill(mydb)
		if err != nil {
			return err
		}
		var bp_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				bp_info[cols[i]] = row[i]
			}
		}
		cols, data, err = GetRedoInfo(mydb)
		if err != nil {
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
			return err
		}
		var ahi_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				ahi_info[cols[i]] = row[i]
			}
		}
		graph_pct, _ := strconv.Atoi(bp_info["BufferPoolFull"])
		bp_graph.Percent(graph_pct)
		chkpt_pct, _ := strconv.Atoi(redo_info["CheckpointAgeInt"])
		redo_graph.Percent(chkpt_pct)
		graph_pct, _ = strconv.Atoi(ahi_info["AHIRatioInt"])
		ahi_graph.Percent(graph_pct)
		uptime_sec, _ := strconv.Atoi(redo_info["Uptime"])
		top_window.Reset()
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
			top_window.Write(PrintLabel("InnodDB Log File Size"))
			top_window.Write(fmt.Sprintf("%-10v", redo_info["InnoDBLogFileSize"]))
			top_window.Write("\n")
			top_window.Write(PrintLabel("Num InnoDB Log File"))
			top_window.Write(fmt.Sprintf("%-10v", redo_info["NbFiles"]))
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
						container.Border(linestyle.Light),
						container.ID("main_container"),
						container.PlaceWidget(details_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.SplitFixed(14),
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
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("InnoDB Info (<-- <Backspace> to return to Processlist)"))
	top_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))
	details_window.Write("\n\n... To be implemented...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

	quitter := func(k2 *terminalapi.Keyboard) {
		if k2.Key == keyboard.KeyEsc || k2.Key == keyboard.KeyCtrlC {
			k = k2.Key
			cancel()
			return
		} else if k2.Key == keyboard.KeyBackspace2 {
			cancel()
			return
		}
	}
	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		return k, err
	}
	return k, nil
}
