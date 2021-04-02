package innotop

import (
	"fmt"

	"context"

	"github.com/alexeyco/simpletable"

	"github.com/mum4k/termdash/terminal/tcell"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

func TableFromSlice(header []string, contents [][]string, style *simpletable.Style) string {
	table := simpletable.New()
	if len(header) > 0 {
		var cells = make([]*simpletable.Cell, len(header))
		for i, h := range header {
			cells[i] = &simpletable.Cell{
				Align: simpletable.AlignCenter, Text: h,
			}
		}
		table.Header = &simpletable.Header{Cells: cells}
	}
	for _, row := range contents {
		var cells []*simpletable.Cell
		for _, item := range row {
			cells = append(cells, &simpletable.Cell{
				Align: simpletable.AlignLeft,
				Text:  item,
			})
		}
		table.Body.Cells = append(table.Body.Cells, cells)
	}
	if style == nil {
		style = simpletable.StyleDefault
	}
	table.SetStyle(style)
	return table.String()
}

func DisplaySimple(cols []string, data [][]string) {
	table := TableFromSlice(cols, data, nil)
	fmt.Printf("%s\n", table)
}

func Display(cols []string, data [][]string) {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	ctx, cancel := context.WithCancel(context.Background())
	borderless, err := text.New()
	if err != nil {
		panic(err)
	}
	if err := borderless.Write("InnoTop Go"); err != nil {
		panic(err)
	}

	c, err := container.New(
		t,
		container.Border(linestyle.Light),
		container.BorderTitle("PRESS Q TO QUIT"),
		container.PlaceWidget(borderless),
	)
	if err != nil {
		panic(err)
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter)); err != nil {
		panic(err)
	}
}
