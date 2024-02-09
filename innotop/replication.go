package innotop

import (
	"database/sql"
	"context"
	"fmt"
	"time"
	"strconv"

	"github.com/mum4k/termdash"
  "github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
)

func GetReplicaStatus(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SHOW REPLICA STATUS;`

	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}

	return cols, data, nil
}

func GetSourceStatus(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SHOW REPLICAS;`

	rows, err := db.Query(mydb, stmt)
	if err != nil {
		return nil, nil, err
	}
	cols, data, err := db.GetData(rows)
	if err != nil {
		return nil, nil, err
	}

	return cols, data, nil
}

func refresh_replication_info(t *tcell.Terminal, cancel context.CancelFunc, ctx context.Context, interval time.Duration, fn func() error) {
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

func convertSecondsToDuration(seconds float64) string {
	duration := time.Duration(seconds) * time.Second

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	secondsRemain := int(duration.Seconds()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%d days", days)
	case hours > 0:
		return fmt.Sprintf("%d hours", hours)
	case minutes > 0:
		return fmt.Sprintf("%d minutes", minutes)
	default:
		return fmt.Sprintf("%d seconds", secondsRemain)
	}
}

func DisplayReplication(mydb *sql.DB, c *container.Container, t *tcell.Terminal) (keyboard.Key, error) {
	ctxmem, cancel := context.WithCancel(context.Background())
	k := keyboard.KeyBackspace2
	sourceStatusWidget, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}
	replicationStatusWidget, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	var replica_info = make(map[string]string)
	var source_info = make(map[string]string)

	// graph replication lag
	// as two sparkline graphs on top of each other, as linechart does not seem to support labelled lines
	replication_lag_graph, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.YAxisFormattedValues(convertSecondsToDuration),
		linechart.YAxisCustomScale(0, 5),
	)
	if err != nil {
		cancel()
		return k, err
	}

	var graphValues map[string][]float64

	// Initialize the map
	graphValues = make(map[string][]float64)

	go refresh_replication_info(t, cancel, ctxmem, 1*time.Second, func() error {
		cols, data, err := GetReplicaStatus(mydb)
		if err != nil {
			return err
		}

		sourceCols, source_data, err := GetSourceStatus(mydb)
		if err != nil {
			return err
		}

		sourceStatusWidget.Reset()
		sourceStatusWidget.Write("\n")

		if source_data == nil {
			sourceStatusWidget.Write("This host does not have any replicas configured")
			sourceStatusWidget.Write("\n")
		} else {

			sourceStatusWidget.Write(fmt.Sprintf("%-10v %-35v %-10v %-36v\n", "Server ID", "Host", "Source ID", "Replica UUID"), text.WriteCellOpts(cell.Bold()))

			for _, row := range source_data {
				for i := 0; i < len(row); i++ {
					source_info[sourceCols[i]] = row[i]
				}

				sourceStatusWidget.Write(fmt.Sprintf("%-10v %-35v %-10v %-36v\n",
					ChunkString(source_info["Server_Id"], 10),
					ChunkString(source_info["Host"], 35),
					ChunkString(source_info["Source_Id"], 10),
					ChunkString(source_info["Replica_UUID"], 36),
				))
				}
		}

		replicationStatusWidget.Reset()
		replicationStatusWidget.Write("\n")

		if data == nil {
			replicationStatusWidget.Write("This host has no replication configured")
			replicationStatusWidget.Write("\n")
		} else {
			for _, row := range data {
				for i := 0; i < len(row); i++ {
					replica_info[cols[i]] = row[i]
				}

				// Update the replication lag linechart
				lag_int, _ := strconv.Atoi(replica_info["Seconds_Behind_Source"])
				lag := float64(lag_int)

				graphValues[replica_info["Channel_Name"]] = append(graphValues[replica_info["Channel_Name"]], lag)

				// write the replication status
				replicationStatusWidget.Write(PrintLabel("ChannelName"))
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Channel_Name"]))
				replicationStatusWidget.Write("\n")
				replicationStatusWidget.Write(PrintLabel("Source Host"))
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Source_Host"]))
				replicationStatusWidget.Write("\n")
				replicationStatusWidget.Write(PrintLabel("Replica IO Running"))

				var io_colour int
				replica_io_running, _ := replica_info["Replica_IO_Running"]
				switch {
				case replica_io_running == "No":
					io_colour = 9
					default:
						io_colour = 2
				}
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Replica_IO_Running"]), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(io_colour))))
				replicationStatusWidget.Write("\n")
				replicationStatusWidget.Write(PrintLabel("Replica SQL Running"))
				var sql_colour int
				replica_sql_running, _ := replica_info["Replica_SQL_Running"]
				switch {
				case replica_sql_running == "No":
					sql_colour = 9
					default:
						sql_colour = 2
				}
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Replica_SQL_Running"]), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(sql_colour))))
				replicationStatusWidget.Write("\n")
				replicationStatusWidget.Write(PrintLabel("Replica SQL Running State"))
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Replica_SQL_Running_State"]))
				replicationStatusWidget.Write("\n")
				var secs_behind_colour int
				seconds_behind_source, _ := replica_info["Seconds_Behind_Source"]
				switch {
				case seconds_behind_source != "0":
					secs_behind_colour = 9
					default:
						secs_behind_colour = 2
				}
				replicationStatusWidget.Write(PrintLabel("Seconds Behind Source"))
				replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Seconds_Behind_Source"]), text.WriteCellOpts(cell.FgColor(cell.ColorNumber(secs_behind_colour))))
				replicationStatusWidget.Write("\n")
				gtid_replication, err := strconv.Atoi(string(replica_info["Auto_Position"]))
				if err != nil {
					fmt.Println(replica_info["Auto_Position"])
					fmt.Println(string(replica_info["Auto_Position"]))
					fmt.Println(err)
					return err
				}
				switch {
				case gtid_replication == 0:
					replicationStatusWidget.Write(PrintLabel("Source Log File"))
					replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Source_Log_File"]))
					replicationStatusWidget.Write("\n")
					replicationStatusWidget.Write(PrintLabel("Read Source Log Pos"))
					replicationStatusWidget.Write(fmt.Sprintf("%s", replica_info["Read_Source_Log_Pos"]))
					replicationStatusWidget.Write("\n")
					skip_counter, _ := strconv.Atoi(replica_info["Skip_Counter"])
					replicationStatusWidget.Write(PrintLabel("Skip Counter"))
					replicationStatusWidget.Write(fmt.Sprintf("%d", skip_counter))
					replicationStatusWidget.Write("\n")
					default:
						retrieved_gtid_set, _ := strconv.Atoi(replica_info["Retrieved_Gtid_Set"])
						executed_gtid_sets, _ := strconv.Atoi(replica_info["Executed_Gtid_Sets"])
						replicationStatusWidget.Write(PrintLabel("Retrieved Gtid Set"))
						replicationStatusWidget.Write(fmt.Sprintf("%d", retrieved_gtid_set))
						replicationStatusWidget.Write("\n")
						replicationStatusWidget.Write(PrintLabel("Executed Gtid Set"))
						replicationStatusWidget.Write(fmt.Sprintf("%d", executed_gtid_sets))
						replicationStatusWidget.Write("\n\n")
				}
			}

			colors := []int{33, 44, 55}
			var iterator int

			for channel_name, values := range graphValues {
				color := linechart.SeriesCellOpts(cell.FgColor(cell.ColorNumber(colors[iterator])))
				replication_lag_graph.Series(channel_name, values, color)

				iterator += 1

				// In case we have more sources than colors, we reset the iterator to
				// prevent an out-of-bounds issue. This might lead to duplicate colors,
				// but will prevent the app from crashing.
				if iterator >= len(colors) {
					iterator = 0
				}
			}
		}

		return nil
	})

	c.Update("dyn_top_container",
		container.SplitHorizontal(
			container.Top(
				container.Border(linestyle.Light),
				container.ID("source_status_container"),
				container.PlaceWidget(sourceStatusWidget),
				container.FocusedColor(cell.ColorNumber(15)),
			),
			container.Bottom(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.ID("replication_status_container"),
						container.PlaceWidget(replicationStatusWidget),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.ID("replication_lag_graph"),
						container.FocusedColor(cell.ColorNumber(15)),
						container.PlaceWidget(replication_lag_graph),
					),
					container.SplitPercent(65),
				),
			),
		),
	)
	c.Update("replication_status_container", container.Focused())
	c.Update("source_status_container", container.BorderTitle("Source Status (<-- <Backspace> to return to Processlist)"))
	c.Update("replication_status_container", container.BorderTitle("Replication Status"))
	c.Update("replication_lag_graph", container.BorderTitle("Replication Lag"))
	replicationStatusWidget.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

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
