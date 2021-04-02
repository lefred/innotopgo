package innotop

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

const redrawInterval = 1000 * time.Millisecond

func DisplayProcesslist(cols []string, data [][]string) {

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

	header := fmt.Sprintf("%-7v %-5v %-5v %-5v %-15v %-20v %-12v %-10v %-10v %-65v\n", "Cmd", "Thd", "Conn", "Pid", "State", "User", "Db", "Time", "Lock Time", "Query")
	if err := borderless.Write(header, text.WriteCellOpts(cell.Bold())); err != nil {
		panic(err)
	}
	for _, row := range data {
		line := fmt.Sprintf("%-7v %-5v %-5v %-5v %-15v %-20v %-12v %-10v %-10v %-65v\n", row[0], row[1], row[2], row[3], row[4], row[5], row[6], row[8], row[9], row[7])
		borderless.Write(line)
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

	if err := termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		panic(err)
	}
}
