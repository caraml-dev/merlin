package table

import (
	"testing"

	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
)

// Add old update column function in table_test.go
func (t *Table) oldUpdateColumnsRaw(columnValues map[string]interface{}) error {
	origColumns := t.Columns()

	newColumns, err := createSeries(columnValues, t.NRow())
	if err != nil {
		return err
	}

	combinedColumns := make([]*series.Series, 0)
	combinedColumns = append(combinedColumns, newColumns...)
	for _, origColumn := range origColumns {
		origColumnName := origColumn.Series().Name
		_, ok := columnValues[origColumnName]
		if ok {
			continue
		}
		combinedColumns = append(combinedColumns, origColumn)
	}

	newT := New(combinedColumns...)
	t.dataFrame = newT.dataFrame
	return nil
}

// goos: darwin
// goarch: amd64
// pkg: github.com/caraml-dev/merlin/pkg/transformer/types/table
// Benchmark_oldUpdateColumnsRaw-12    	  175756	      6568 ns/op
// Benchmark_UpdateColumnsRaw-12       	  181879	      6454 ns/op
// PASS
// ok  	github.com/caraml-dev/merlin/pkg/transformer/types/table	2.728s
// Add two new benchmark functions. For the old one and new one
func Benchmark_oldUpdateColumnsRaw(b *testing.B) { // old
	table := New(
		series.New([]int{1, 2, 3}, series.Int, "col1"),
		series.New([]int{11, 22, 33}, series.Int, "col2"),
		series.New([]int{111, 222, 333}, series.Int, "col3"),
	)
	newColumns := map[string]interface{}{
		"col4": []int{12, 14, 16},
		"col2": []int{58, 55, 5},
		"col5": []float64{3.14, 4.26, 9.88},
		"col1": []int{88, 168, -222},
	}

	for i := 0; i < b.N; i++ {
		err := table.oldUpdateColumnsRaw(newColumns)
		if err != nil {
			panic(err)
		}
	}
}

func Benchmark_UpdateColumnsRaw(b *testing.B) { // new
	table := New(
		series.New([]int{1, 2, 3}, series.Int, "col1"),
		series.New([]int{11, 22, 33}, series.Int, "col2"),
		series.New([]int{111, 222, 333}, series.Int, "col3"),
	)
	newColumns := map[string]interface{}{
		"col4": []int{12, 14, 16},
		"col2": []int{58, 55, 5},
		"col5": []float64{3.14, 4.26, 9.88},
		"col1": []int{88, 168, -222},
	}

	for i := 0; i < b.N; i++ {
		err := table.UpdateColumnsRaw(newColumns)
		if err != nil {
			panic(err)
		}
	}
}
