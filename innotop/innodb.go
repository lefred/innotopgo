package innotop

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lefred/innotopgo/db"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
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

func GetBPFill(mydb *sql.DB) ([]string, [][]string, error) {
	stmt := `SELECT sleep(2) as meh, ROUND(A.num * 100.0 / B.num)  BufferPoolFull, BP_Size, BP_instances    
	         FROM (     
				    SELECT variable_value num FROM performance_schema.global_status
					WHERE variable_name = 'Innodb_buffer_pool_pages_data') A,       
				  (
					SELECT variable_value num FROM performance_schema.global_status       
					WHERE variable_name = 'Innodb_buffer_pool_pages_total') B, 
				  (
					SELECT format_bytes(variable_value) as BP_Size 
					FROM performance_schema.global_variables 
					WHERE variable_name = 'innodb_buffer_pool_size') C, 
				  (
					SELECT variable_value as BP_instances 
					FROM performance_schema.global_variables 
					WHERE variable_name = 'innodb_buffer_pool_instances') D
	`
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

func DisplayInnoDB(ctx context.Context, mydb *sql.DB, c *container.Container) error {
	details_window, err := text.New()
	if err != nil {
		return err
	}
	bp_graph, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorNumber(31))),
		donut.Label("Buffer Pool %", cell.FgColor(cell.ColorNumber(31))),
	)
	if err != nil {
		return err
	}

	top_window, err := text.New(text.WrapAtWords())

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
		bp_pct, _ := strconv.Atoi(bp_info["BufferPoolFull"])
		bp_graph.Percent(bp_pct)
		top_window.Reset()
		top_window.Write(fmt.Sprintf("  Buffer Pool Size: %-10v\n", bp_info["BP_Size"]))
		top_window.Write(fmt.Sprintf("  Buffer Instances: %-10v\n", bp_info["BP_instances"]))
		return nil
	})

	c.Update("dyn_top_container",
		container.SplitHorizontal(
			container.Top(
				container.SplitVertical(
					container.Left(
						container.Border(linestyle.Light),
						container.ID("top_container"),
						container.PlaceWidget(top_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.Right(
						container.SplitVertical(
							container.Left(
								container.Border(linestyle.Light),
								container.ID("top_left_graph"),
								container.FocusedColor(cell.ColorNumber(15)),
								//container.PlaceWidget(tlg),
							),
							container.Right(
								container.Border(linestyle.Light),
								container.ID("top_right_graph"),
								container.FocusedColor(cell.ColorNumber(15)),
								container.PlaceWidget(bp_graph),
							),
							container.SplitPercent(50),
						),
					),
					container.SplitPercent(60),
				),
			),
			container.Bottom(
				container.Border(linestyle.Light),
				container.ID("main_container"),
				container.PlaceWidget(details_window),
				container.FocusedColor(cell.ColorNumber(15)),
			),
			container.SplitFixed(10),
		),
	)
	c.Update("bottom_container", container.Clear())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.Focused())
	c.Update("main_container", container.BorderTitle("InnDB Info (<-- <Backspace> to return to Processlist)"))
	details_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

	return nil
}
