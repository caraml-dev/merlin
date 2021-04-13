package table

import (
	"testing"

	"github.com/go-gota/gota/dataframe"
	gotaSeries "github.com/go-gota/gota/series"
	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/types/series"
)

func TestTable_New(t *testing.T) {
	table := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col"),
		series.New([]int{1111, 2222}, series.Int, "int32_col"),
		series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
		series.New([]float64{1111, 2222}, series.Float, "float32_col"),
		series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
		series.New([]bool{true, false}, series.Bool, "bool_col"),
	)
	gotaDataFrame := dataframe.New(
		gotaSeries.New([]string{"1111", "2222"}, gotaSeries.String, "string_col"),
		gotaSeries.New([]int{1111, 2222}, gotaSeries.Int, "int32_col"),
		gotaSeries.New([]int{1111111111, 2222222222}, gotaSeries.Int, "int64_col"),
		gotaSeries.New([]float64{1111, 2222}, gotaSeries.Float, "float32_col"),
		gotaSeries.New([]float64{11111111111.1111, 22222222222.2222}, gotaSeries.Float, "float64_col"),
		gotaSeries.New([]bool{true, false}, gotaSeries.Bool, "bool_col"),
	)

	assert.Equal(t, gotaDataFrame, *table.DataFrame())

	table2 := NewTable(&gotaDataFrame)
	assert.Equal(t, gotaDataFrame, *table2.DataFrame())
}

func TestTable_Col(t *testing.T) {
	table := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col"),
	)

	col, err := table.Col("string_col")
	assert.NoError(t, err)
	assert.Equal(t, series.New([]string{"1111", "2222"}, series.String, "string_col"), col)

	col, err = table.Col("col_not_exists")
	assert.Error(t, err, "unknown column name")
}

func TestTable_Row(t *testing.T) {
	table := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col"),
	)

	row, err := table.Row(0)
	assert.NoError(t, err)
	assert.Equal(t, table.DataFrame().Subset(0), *row.DataFrame())

	row, err = table.Row(2)
	assert.Error(t, err)
	assert.Equal(t, "invalid row number, expected: 0 <= row < 2, got: 2", err.Error())
}

func TestTable_Copy(t *testing.T) {
	table := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_2"),
	)

	tableCopy := table.Copy()
	assert.Equal(t, table.DataFrame(), tableCopy.DataFrame())

	// assert that modifying copy won't affect the original
	df := tableCopy.DataFrame().Drop("string_col_2")
	assert.ElementsMatch(t, []string{"string_col"}, df.Names())
	assert.ElementsMatch(t, []string{"string_col", "string_col_2"}, table.DataFrame().Names())
}

func TestTable_ConcatColumn(t *testing.T) {
	table1 := New(
		series.New([]string{"1111", "1111"}, series.String, "string_col"),
		series.New([]string{"11", "22"}, series.String, "string_col_2"),
	)

	table2 := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col_2"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_3"),
	)

	table3, err := table1.ConcatColumn(table2)
	assert.NoError(t, err)
	assert.Equal(t, New(
		series.New([]string{"1111", "1111"}, series.String, "string_col"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_2"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_3"),
	).DataFrame(), table3.DataFrame())

	longerTable := New(
		series.New([]string{"1111", "1111", "1111"}, series.String, "string_col_3"),
	)

	table3, err = table1.ConcatColumn(longerTable)
	assert.Error(t, err)
	assert.Nil(t, table3)
}
