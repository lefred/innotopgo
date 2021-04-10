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
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
)

func refresh_memory_info(t *tcell.Terminal, cancel context.CancelFunc, ctx context.Context, interval time.Duration, fn func() error) {
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

func DisplayMemory(mydb *sql.DB, c *container.Container, t *tcell.Terminal) (keyboard.Key, error) {
	ctxmem, cancel := context.WithCancel(context.Background())
	k := keyboard.KeyBackspace2
	tot_mem_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}
	user_mem_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	mem_graph, err := sparkline.New(
		sparkline.Color(cell.ColorBlue),
	)
	if err != nil {
		cancel()
		return k, err
	}

	temp_window, err := text.New()
	if err != nil {
		cancel()
		return k, err
	}

	var prev_mem_info = make(map[string]string)

	go refresh_memory_info(t, cancel, ctxmem, 1*time.Second, func() error {
		cols, data, err := GetTempMem(mydb)
		if err != nil {
			cancel()
			return err
		}
		var mem_info = make(map[string]string)
		for _, row := range data {
			for i := 0; i < len(row); i++ {
				mem_info[cols[i]] = row[i]
			}
		}
		_, data, err = GetTempAlloc(mydb)
		if err != nil {
			cancel()
			return err
		}
		var mem_alloc = make(map[string][]string)
		for _, row := range data {
			mem_alloc[row[0]] = []string{row[1], row[2]}
		}
		_, data, err = GetUserMemAlloc(mydb)
		if err != nil {
			cancel()
			return err
		}
		var user_mem_alloc = make(map[string][]string)
		for _, row := range data {
			user_mem_alloc[row[0]] = []string{row[1], row[2]}
		}

		_, code_mem_alloc, err := GetCodeMemAlloc(mydb)
		if err != nil {
			cancel()
			return err
		}

		tot_mem_alloc, _ := strconv.Atoi(mem_info["TotalAllocatedNum"])
		mem_graph.Add(([]int{tot_mem_alloc}))
		// Do the work and printing here
		uptime_sec, _ := strconv.Atoi(mem_info["Uptime"])
		tot_mem_window.Reset()
		tot_mem_window.Write("\n")
		tot_mem_window.Write(PrintLabel("Total Memory Allocated"))
		tot_mem_window.Write(fmt.Sprintf("%-10v", mem_info["TotalAllocated"]))
		tot_mem_window.Write(PrintLabel("Uptime"))
		tot_mem_window.Write(fmt.Sprintf("%-10v", (time.Duration(uptime_sec) * time.Second)))
		tot_mem_window.Write("\n\n")

		tot_mem_window.Write(PrintLabel("Code Area"))
		tot_mem_window.Write(PrintLabel("Memory Allocation"))
		tot_mem_window.Write("\n\n")

		for _, row := range code_mem_alloc {
			tot_mem_window.Write(fmt.Sprintf("%27v %28v", row[0], row[1]))
			tot_mem_window.Write("\n")
		}

		temp_window.Reset()
		temp_window.Write(("\n"))
		// Get the value of temp tables to
		temp_tbl := GetValue(prev_mem_info, mem_info, "TempTables")
		temp_tbl_disk := GetValue(prev_mem_info, mem_info, "TempTablesDisk")

		temp_window.Write("RAM:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
		temp_window.Write(("\n\n"))
		temp_window.Write(PrintLabel("Current", 0))
		temp_window.Write(fmt.Sprintf("%6v", temp_tbl))
		temp_window.Write(("\n"))
		if len(mem_alloc["memory/temptable/physical_ram"]) > 0 {
			temp_window.Write(fmt.Sprintf("  current allocation: %v", mem_alloc["memory/temptable/physical_ram"][0]))
			temp_window.Write(("\n"))
			temp_window.Write(fmt.Sprintf("     high allocation: %v", mem_alloc["memory/temptable/physical_ram"][1]))
		}

		temp_window.Write(("\n\n"))
		temp_window.Write("DISK:", text.WriteCellOpts(cell.Bold(), cell.Underline()))
		temp_window.Write(("\n\n"))
		temp_window.Write(PrintLabel("Current", 0))
		temp_window.Write(fmt.Sprintf("%6v", temp_tbl_disk))
		temp_window.Write(("\n"))
		if len(mem_alloc["memory/temptable/physical_disk"]) > 0 {
			temp_window.Write(fmt.Sprintf("  current allocation: %v", mem_alloc["memory/temptable/physical_disk"][0]))
			temp_window.Write(("\n"))
			temp_window.Write(fmt.Sprintf("     high allocation: %v", mem_alloc["memory/temptable/physical_disk"][1]))
		}
		temp_window.Write(("\n\n"))
		temp_window.Write(PrintLabel("      Memory Engine", 0))
		temp_window.Write(fmt.Sprintf("%v", mem_info["TempMemStorageEngine"]))
		temp_window.Write(("\n"))
		if mem_info["TempMemStorageEngine"] == "TempTable" {
			temp_window.Write(PrintLabel("       Max RAM Size", 0))
			temp_window.Write(fmt.Sprintf("%v", mem_info["TempMaxRam"]))
			temp_window.Write(("\n"))
			temp_window.Write(PrintLabel("     Temp uses Nmap", 0))
			temp_window.Write(fmt.Sprintf("%v", mem_info["TempUseNmap"]))
			temp_window.Write(("\n"))
			temp_window.Write(PrintLabel("      Temp Max Nmap", 0))
			temp_window.Write(fmt.Sprintf("%v", mem_info["TempMaxNmap"]))
		} else if mem_info["TempMemStorageEngine"] == "MEMORY" {
			temp_window.Write(PrintLabel(" Max RAM Table Size", 0))
			memory_table_size, _ := strconv.Atoi(mem_info["TmpTableSize"])
			heap_table_size, _ := strconv.Atoi(mem_info["MaxHeapTableSize"])
			tmp_table_size := memory_table_size
			if heap_table_size < tmp_table_size {
				tmp_table_size = heap_table_size
			}
			temp_window.Write(fmt.Sprintf("%v", FormatBytes(tmp_table_size)))
		}
		temp_window.Write(("\n\n"))
		temp_window.Write(PrintLabel("Temp Tbl Disk Ratio", 0))
		temp_window.Write(fmt.Sprintf("%v%%", mem_info["TempTablesDiskRatio"]))

		prev_mem_info = mem_info

		user_mem_window.Reset()
		user_mem_window.Write("\n")
		user_mem_window.Write(PrintLabel("User", 0))
		user_mem_window.Write(PrintLabel("Current Mem"))
		user_mem_window.Write(PrintLabel("Max Mem"))
		user_mem_window.Write("\n\n")
		for i := range user_mem_alloc {
			user_mem_window.Write(fmt.Sprintf("%-22v", i))
			user_mem_window.Write(fmt.Sprintf("%-33v", user_mem_alloc[i][0]))
			user_mem_window.Write(fmt.Sprintf("%-22v", user_mem_alloc[i][1]))
			user_mem_window.Write("\n")
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
						container.PlaceWidget(tot_mem_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.ID("user_container"),
						container.PlaceWidget(user_mem_window),
						container.FocusedColor(cell.ColorNumber(15)),
					),
					container.SplitPercent(65),
				),
			),
			container.Right(
				container.SplitHorizontal(
					container.Top(
						container.Border(linestyle.Light),
						container.ID("memory_alloc_container"),
						container.FocusedColor(cell.ColorNumber(15)),
						container.PlaceWidget(mem_graph),
					),
					container.Bottom(
						container.Border(linestyle.Light),
						container.ID("left_temp"),
						container.FocusedColor(cell.ColorNumber(15)),
						container.PlaceWidget(temp_window),
					),
					container.SplitPercent(30),
				),
			),
			container.SplitFixed(85),
		),
	)
	c.Update("user_container", container.Focused())
	c.Update("top_container", container.BorderTitle("Memory Allocation (<-- <Backspace> to return to Processlist)"))
	c.Update("user_container", container.BorderTitle("Users Memory Allocation"))
	c.Update("temp_container", container.BorderTitle("Temporary Tables"))
	c.Update("memory_alloc_container", container.BorderTitle("Total Memory Allocated"))
	tot_mem_window.Write("\n\n... please wait...", text.WriteCellOpts(cell.FgColor(cell.ColorNumber(6)), cell.Italic()))

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
