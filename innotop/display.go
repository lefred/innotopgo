package innotop

import (
	"fmt"

	"github.com/alexeyco/simpletable"
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
