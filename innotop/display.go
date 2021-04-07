package innotop

import (
	"fmt"
	"strings"

	"github.com/alexeyco/simpletable"
	"github.com/mum4k/termdash/cell"
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

func ChunkString(s string, chunkSize int) string {
	if chunkSize >= len(s) {
		return s
	}
	var chunks []string
	chunk := make([]rune, chunkSize)
	len := 0
	for _, r := range s {
		chunk[len] = r
		len++
		if len == chunkSize {
			chunks = append(chunks, string(chunk))
			len = 0
		}
	}
	if len > 0 {
		chunks = append(chunks, string(chunk[:len]))
	}
	return chunks[0]
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
