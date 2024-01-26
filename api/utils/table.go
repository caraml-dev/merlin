package utils

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

func LogTable(headers []string, rows [][]string) string {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)

	headerRow := table.Row{}
	for _, v := range headers {
		headerRow = append(headerRow, v)
	}
	t.AppendHeader(headerRow)

	for _, v := range rows {
		tableRow := table.Row{}
		for _, cell := range v {
			tableRow = append(tableRow, cell)
		}
		t.AppendRow(tableRow)
	}

	return t.Render()
}
