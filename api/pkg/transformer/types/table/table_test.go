package table

import (
	"errors"
	"testing"

	"github.com/go-gota/gota/dataframe"
	gotaSeries "github.com/go-gota/gota/series"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/stretchr/testify/assert"
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

	col, err := table.GetColumn("string_col")
	assert.NoError(t, err)
	assert.Equal(t, series.New([]string{"1111", "2222"}, series.String, "string_col"), col)

	col, err = table.GetColumn("col_not_exists")
	assert.Error(t, err, "unknown column name")
}

func TestTable_Row(t *testing.T) {
	table := New(
		series.New([]string{"1111", "2222"}, series.String, "string_col"),
	)

	row, err := table.GetRow(0)
	assert.NoError(t, err)
	assert.Equal(t, table.DataFrame().Subset(0), *row.DataFrame())

	row, err = table.GetRow(2)
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

	table3, err := table1.Concat(table2)
	assert.NoError(t, err)
	assert.Equal(t, New(
		series.New([]string{"1111", "1111"}, series.String, "string_col"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_2"),
		series.New([]string{"1111", "2222"}, series.String, "string_col_3"),
	).DataFrame(), table3.DataFrame())

	longerTable := New(
		series.New([]string{"1111", "1111", "1111"}, series.String, "string_col_3"),
	)

	table3, err = table1.Concat(longerTable)
	assert.Error(t, err)
	assert.Nil(t, table3)
}

func TestTable_DropColumns(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name      string
		srcTable  *Table
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name: "success: all columns exists",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columns: []string{"string_col", "int32_col"},
			},
			expTable: New(
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			wantError: false,
		},
		{
			name: "success: drop zero columns",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columns: []string{},
			},
			expTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			wantError: false,
		},
		{
			name: "failed: drop unknown columns",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columns: []string{"unknown_columns"},
			},
			wantError: true,
			expError:  errors.New("can't select columns: can't select columns: column name \"unknown_columns\" not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcTable.DropColumns(tt.args.columns)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, tt.srcTable)
		})
	}
}

func TestTable_SelectColumns(t *testing.T) {
	type args struct {
		columns []string
	}
	tests := []struct {
		name      string
		srcTable  *Table
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name: "success: all columns exists",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columns: []string{"int32_col", "string_col"},
			},
			expTable: New(
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
			),
			wantError: false,
		},
		{
			name: "error: unknown columns",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columns: []string{"int32_col", "unknown_column"},
			},
			wantError: true,
			expError:  errors.New("can't select columns: can't select columns: column name \"unknown_column\" not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcTable.SelectColumns(tt.args.columns)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, tt.srcTable)
		})
	}
}

func TestTable_RenameColumns(t *testing.T) {
	type args struct {
		columnMap map[string]string
	}
	tests := []struct {
		name      string
		srcTable  *Table
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name: "success: rename one column",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columnMap: map[string]string{
					"string_col": "new_string_col",
				},
			},
			expTable: New(
				series.New([]string{"1111", "2222"}, series.String, "new_string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			wantError: false,
		},
		{
			name: "success: rename multiple columns",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columnMap: map[string]string{
					"string_col": "new_string_col",
					"int32_col":  "new_int32_col",
					"int64_col":  "new_int64_col",
				},
			},
			expTable: New(
				series.New([]string{"1111", "2222"}, series.String, "new_string_col"),
				series.New([]int{1111, 2222}, series.Int, "new_int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "new_int64_col"),
			),
			wantError: false,
		},
		{
			name: "error: rename unknown columns",
			srcTable: New(
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
			),
			args: args{
				columnMap: map[string]string{
					"unknown_column": "new_unknown_column",
				},
			},
			expTable: New(
				series.New([]string{"1111", "2222"}, series.String, "new_string_col"),
				series.New([]int{1111, 2222}, series.Int, "new_int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "new_int64_col"),
			),
			wantError: true,
			expError:  errors.New("unable to rename column: unknown column: unknown_column"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcTable.RenameColumns(tt.args.columnMap)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, tt.srcTable)
		})
	}
}

func TestTable_Sort(t *testing.T) {
	type args struct {
		sortRules []*spec.SortColumnRule
	}
	tests := []struct {
		name      string
		srcTable  *Table
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name: "success: sort by one column ascending",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "col1",
						Order:  spec.SortOrder_ASC,
					},
				},
			},
			expTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "success: sort by one column descending",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "col1",
						Order:  spec.SortOrder_DESC,
					},
				},
			},
			expTable: New(
				series.New([]int{3, 2, 1}, series.Int, "col1"),
				series.New([]int{33, 22, 11}, series.Int, "col2"),
				series.New([]int{333, 222, 111}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "success: sort by one column descending",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "col1",
						Order:  spec.SortOrder_DESC,
					},
				},
			},
			expTable: New(
				series.New([]int{3, 2, 1}, series.Int, "col1"),
				series.New([]int{33, 22, 11}, series.Int, "col2"),
				series.New([]int{333, 222, 111}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "success: sort by 2 columns both descending",
			srcTable: New(
				series.New([]int{2, 2, 3}, series.Int, "col1"),
				series.New([]int{22, 11, 33}, series.Int, "col2"),
				series.New([]int{222, 222, 333}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "col1",
						Order:  spec.SortOrder_DESC,
					},
					{
						Column: "col2",
						Order:  spec.SortOrder_DESC,
					},
				},
			},
			expTable: New(
				series.New([]int{3, 2, 2}, series.Int, "col1"),
				series.New([]int{33, 22, 11}, series.Int, "col2"),
				series.New([]int{333, 222, 222}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "success: sort by 2 columns both ascending and descending",
			srcTable: New(
				series.New([]int{3, 2, 2}, series.Int, "col1"),
				series.New([]int{33, 11, 22}, series.Int, "col2"),
				series.New([]int{333, 222, 222}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "col1",
						Order:  spec.SortOrder_ASC,
					},
					{
						Column: "col2",
						Order:  spec.SortOrder_DESC,
					},
				},
			},
			expTable: New(
				series.New([]int{2, 2, 3}, series.Int, "col1"),
				series.New([]int{22, 11, 33}, series.Int, "col2"),
				series.New([]int{222, 222, 333}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "error: unknown column",
			srcTable: New(
				series.New([]int{3, 2, 2}, series.Int, "col1"),
				series.New([]int{33, 11, 22}, series.Int, "col2"),
				series.New([]int{333, 222, 222}, series.Int, "col3"),
			),
			args: args{
				sortRules: []*spec.SortColumnRule{
					{
						Column: "unknown_column",
						Order:  spec.SortOrder_ASC,
					},
				},
			},
			wantError: true,
			expError:  errors.New("colname unknown_column doesn't exist"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcTable.Sort(tt.args.sortRules)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, tt.srcTable)
		})
	}
}

func TestTable_UpdateColumnsRaw(t *testing.T) {
	type args struct {
		columnValues map[string]interface{}
	}
	tests := []struct {
		name      string
		srcTable  *Table
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name: "success: update column inplace",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				columnValues: map[string]interface{}{
					"col2": []int{2, 4, 6},
				},
			},
			expTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{2, 4, 6}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			wantError: false,
		},
		{
			name: "success: append new col to end",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				columnValues: map[string]interface{}{
					"col4": []int{12, 14, 16},
				},
			},
			expTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
				series.New([]int{12, 14, 16}, series.Int, "col4"),
			),
			wantError: false,
		},
		{
			name: "success: append new col to end and update existing",
			srcTable: New(
				series.New([]int{1, 2, 3}, series.Int, "col1"),
				series.New([]int{11, 22, 33}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
			),
			args: args{
				columnValues: map[string]interface{}{
					"col4": []int{12, 14, 16},
					"col2": []int{58, 55, 5},
					"col5": []float64{3.14, 4.26, 9.88},
					"col1": []int{88, 168, -222},
				},
			},
			expTable: New(
				series.New([]int{88, 168, -222}, series.Int, "col1"),
				series.New([]int{58, 55, 5}, series.Int, "col2"),
				series.New([]int{111, 222, 333}, series.Int, "col3"),
				series.New([]int{12, 14, 16}, series.Int, "col4"),
				series.New([]float64{3.14, 4.26, 9.88}, series.Float, "col5"),
			),
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srcTable.UpdateColumnsRaw(tt.args.columnValues)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, tt.srcTable)
		})
	}
}

func TestTable_NewRaw(t *testing.T) {
	tests := []struct {
		name         string
		columnValues map[string]interface{}
		want         *Table
		wantErr      bool
	}{
		{
			name: "basic test",
			columnValues: map[string]interface{}{
				"string_col":  []string{"1111", "2222"},
				"int32_col":   []int{1111, 2222},
				"int64_col":   []int{1111111111, 2222222222},
				"float32_col": []float64{1111, 2222},
				"float64_col": []float64{11111111111.1111, 22222222222.2222},
				"bool_col":    []bool{true, false},
			},
			want: New(
				series.New([]bool{true, false}, series.Bool, "bool_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
			),
		},
		{
			name: "basic test - list value type",
			columnValues: map[string]interface{}{
				"string_list_col":  [][]string{{"1111", "1111"}, {"2222", "2222"}},
				"int32_list_col":   [][]int{{1111, 1111}, {2222, 2222}},
				"int64_list_col":   [][]int64{{1111111111, 1111111111}, {2222222222, 2222222222}},
				"float32_list_col": [][]float32{{1111, 1111}, {2222, 2222}},
				"float64_list_col": [][]float64{{11111111111.1111, 11111111111.1111}, {22222222222.2222, 22222222222.2222}},
				"bool_list_col":    [][]bool{{true, true}, {false, false}},
			},
			want: New(
				series.New([][]bool{{true, true}, {false, false}}, series.BoolList, "bool_list_col"),
				series.New([][]float32{{1111, 1111}, {2222, 2222}}, series.FloatList, "float32_list_col"),
				series.New([][]float64{{11111111111.1111, 11111111111.1111}, {22222222222.2222, 22222222222.2222}}, series.FloatList, "float64_list_col"),
				series.New([][]int{{1111, 1111}, {2222, 2222}}, series.IntList, "int32_list_col"),
				series.New([][]int64{{1111111111, 1111111111}, {2222222222, 2222222222}}, series.IntList, "int64_list_col"),
				series.New([][]string{{"1111", "1111"}, {"2222", "2222"}}, series.StringList, "string_list_col"),
			),
		},
		{
			name: "table from series",
			columnValues: map[string]interface{}{
				"string_col":  series.New([]string{"1111", "2222"}, series.String, "string_col"),
				"int32_col":   series.New([]int{1111, 2222}, series.Int, "int32_col"),
				"int64_col":   series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
				"float32_col": series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				"float64_col": series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				"bool_col":    series.New([]bool{true, false}, series.Bool, "bool_col"),
			},
			want: New(
				series.New([]bool{true, false}, series.Bool, "bool_col"),
				series.New([]float64{1111, 2222}, series.Float, "float32_col"),
				series.New([]float64{11111111111.1111, 22222222222.2222}, series.Float, "float64_col"),
				series.New([]int{1111, 2222}, series.Int, "int32_col"),
				series.New([]int{1111111111, 2222222222}, series.Int, "int64_col"),
				series.New([]string{"1111", "2222"}, series.String, "string_col"),
			),
		},
		{
			name: "table from series of list elements",
			columnValues: map[string]interface{}{
				"string_list_col":  series.New([][]string{{"1111", "1111"}, {"2222", "2222"}}, series.StringList, "string_list_col"),
				"int32_list_col":   series.New([][]int{{1111, 1111}, {2222, 2222}}, series.IntList, "int32_list_col"),
				"int64_list_col":   series.New([][]int64{{1111111111, 1111111111}, {2222222222, 2222222222}}, series.IntList, "int64_list_col"),
				"float32_list_col": series.New([][]float32{{1111, 1111}, {2222, 2222}}, series.FloatList, "float32_list_col"),
				"float64_list_col": series.New([][]float64{{11111111111.1111, 11111111111.1111}, {22222222222.2222, 22222222222.2222}}, series.FloatList, "float64_list_col"),
				"bool_list_col":    series.New([][]bool{{true, true}, {false, false}}, series.BoolList, "bool_list_col"),
			},
			want: New(
				series.New([][]bool{{true, true}, {false, false}}, series.BoolList, "bool_list_col"),
				series.New([][]float32{{1111, 1111}, {2222, 2222}}, series.FloatList, "float32_list_col"),
				series.New([][]float64{{11111111111.1111, 11111111111.1111}, {22222222222.2222, 22222222222.2222}}, series.FloatList, "float64_list_col"),
				series.New([][]int{{1111, 1111}, {2222, 2222}}, series.IntList, "int32_list_col"),
				series.New([][]int64{{1111111111, 1111111111}, {2222222222, 2222222222}}, series.IntList, "int64_list_col"),
				series.New([][]string{{"1111", "1111"}, {"2222", "2222"}}, series.StringList, "string_list_col"),
			),
		},
		{
			name: "test with null values",
			columnValues: map[string]interface{}{
				"string_col": []interface{}{"1111", nil},
				"int32_col":  []interface{}{nil, 2222},
				"int64_col":  []interface{}{nil, 2222222222},
			},
			want: New(
				series.New([]interface{}{nil, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{nil, 2222222222}, series.Int, "int64_col"),
				series.New([]interface{}{"1111", nil}, series.String, "string_col"),
			),
		},
		{
			name: "test with null values - series list elements",
			columnValues: map[string]interface{}{
				"string_list_col": []interface{}{[]string{"1111", "2222"}, []string{"2222", "3333"}, nil},
				"int32_list_col":  []interface{}{[]int32{1111, 2222}, nil, []int32{1, 2}},
				"int64_list_col":  []interface{}{[]int64{1111111111, 2222222222}, nil, []int64{3, 4}},
			},
			want: New(
				series.New([][]interface{}{{1111, 2222}, nil, {1, 2}}, series.IntList, "int32_list_col"),
				series.New([][]interface{}{{1111111111, 2222222222}, nil, {3, 4}}, series.IntList, "int64_list_col"),
				series.New([][]interface{}{{"1111", "2222"}, {"2222", "3333"}, nil}, series.StringList, "string_list_col"),
			),
		},
		{
			name: "test broadcast array",
			columnValues: map[string]interface{}{
				"string_col": []interface{}{"1111"},
				"int32_col":  []interface{}{nil, 2222},
				"int64_col":  []interface{}{nil, 2222222222},
			},
			want: New(
				series.New([]interface{}{nil, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{nil, 2222222222}, series.Int, "int64_col"),
				series.New([]interface{}{"1111", "1111"}, series.String, "string_col"),
			),
		},
		{
			name: "test broadcast array - series list to series single",
			columnValues: map[string]interface{}{
				"string_list_col": [][]string{{"1111", "1111"}},
				"int32_col":       []int32{1111, 2222},
				"int64_col":       []int64{1111111111, 2222222222},
			},
			want: New(
				series.New([]interface{}{1111, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{1111111111, 2222222222}, series.Int, "int64_col"),
				series.New([][]string{{"1111", "1111"}, {"1111", "1111"}}, series.StringList, "string_list_col"),
			),
		},
		{
			name: "test broadcast array - series list to series list",
			columnValues: map[string]interface{}{
				"string_list_col": [][]string{{"1111", "1111"}},
				"int32_list_col":  [][]int{{1, 2}, {3, 4}, {5, 6}},
				"int64_list_col":  [][]int{{10, 100}, {200, 2000}, {333, 444}},
			},
			want: New(
				series.New([][]interface{}{{1, 2}, {3, 4}, {5, 6}}, series.IntList, "int32_list_col"),
				series.New([][]interface{}{{10, 100}, {200, 2000}, {333, 444}}, series.IntList, "int64_list_col"),
				series.New([]interface{}{[]string{"1111", "1111"}, []string{"1111", "1111"}, []string{"1111", "1111"}}, series.StringList, "string_list_col"),
			),
		},
		{
			name: "test broadcast scalar",
			columnValues: map[string]interface{}{
				"string_col": "1111",
				"int32_col":  []interface{}{nil, 2222},
				"int64_col":  []interface{}{nil, 2222222222},
			},
			want: New(
				series.New([]interface{}{nil, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{nil, 2222222222}, series.Int, "int64_col"),
				series.New([]interface{}{"1111", "1111"}, series.String, "string_col"),
			),
		},
		{
			name: "test broadcast series",
			columnValues: map[string]interface{}{
				"string_col": series.New([]interface{}{"1111"}, series.String, "string_col"),
				"int32_col":  []interface{}{nil, 2222},
				"int64_col":  []interface{}{nil, 2222222222},
			},
			want: New(
				series.New([]interface{}{nil, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{nil, 2222222222}, series.Int, "int64_col"),
				series.New([]interface{}{"1111", "1111"}, series.String, "string_col"),
			),
		},
		{
			name: "test broadcast series - list elements",
			columnValues: map[string]interface{}{
				"string_list_col": series.New([][]string{{"1111", "2222"}}, series.StringList, "string_list_col"),
				"int32_col":       []interface{}{nil, 2222},
				"int64_col":       []interface{}{nil, 2222222222},
			},
			want: New(
				series.New([]interface{}{nil, 2222}, series.Int, "int32_col"),
				series.New([]interface{}{nil, 2222222222}, series.Int, "int64_col"),
				series.New([][]string{{"1111", "2222"}, {"1111", "2222"}}, series.StringList, "string_list_col"),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRaw(tt.columnValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRaw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.DataFrame(), got.DataFrame())
		})
	}
}

func TestTable_SliceRow(t *testing.T) {
	tests := []struct {
		name       string
		inputTable *Table
		startIdx   int
		endIdx     int
		want       *Table
		wantErr    bool
		errMessage string
	}{
		{
			name: "start < end, end < table row length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: 1,
			endIdx:   3,
			want: New(
				series.New([]int{2, 3}, series.Int, "A"),
				series.New([]string{"b", "c"}, series.String, "B"),
			),
		},
		{
			name: "start < end, end == table row length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: 1,
			endIdx:   5,
			want: New(
				series.New([]int{2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"b", "c", "d", "e"}, series.String, "B"),
			),
		},
		{
			name: "start < end, end > table row length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   1,
			endIdx:     6,
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "start > end",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   2,
			endIdx:     0,
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "start < 0",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   -1,
			endIdx:     2,
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inputTable.SliceRow(tt.startIdx, tt.endIdx)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMessage)
				return
			}
			assert.Equal(t, tt.want, tt.inputTable)
		})
	}
}

func TestTable_FilterRow(t *testing.T) {
	tests := []struct {
		name       string
		inputTable *Table
		subset     *series.Series
		want       *Table
		wantErr    bool
		errorMsg   string
	}{
		{
			name: "Subset length is same with table row",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
				series.New([]float32{1.1, 2.2, 3.3, 4.4, 5.5}, series.Float, "C"),
			),
			subset:  series.New([]bool{false, true, false, true, false}, series.Bool, ""),
			wantErr: false,
			want: New(
				series.New([]int{2, 4}, series.Int, "A"),
				series.New([]string{"b", "d"}, series.String, "B"),
				series.New([]float32{2.2, 4.4}, series.Float, "C"),
			),
		},
		{
			name: "Subset length is less than table row",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
				series.New([]float32{1.1, 2.2, 3.3, 4.4, 5.5}, series.Float, "C"),
			),
			subset:   series.New([]bool{false, true, true}, series.Bool, ""),
			wantErr:  true,
			errorMsg: "error on series 0: indexing error: index dimensions mismatch",
		},
		{
			name: "Subset length is more than table row",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
				series.New([]float32{1.1, 2.2, 3.3, 4.4, 5.5}, series.Float, "C"),
			),
			subset:   series.New([]bool{false, true, true, true, true, true, true}, series.Bool, ""),
			wantErr:  true,
			errorMsg: "error on series 0: indexing error: index dimensions mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inputTable.FilterRow(tt.subset)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errorMsg)
				return
			}
			assert.Equal(t, tt.want, tt.inputTable)
		})
	}
}

func TestTable_UpdateColumns(t *testing.T) {
	tests := []struct {
		name           string
		inputTable     *Table
		updateColRules []ColumnUpdate
		want           *Table
		wantErr        bool
		errMessage     string
	}{
		{
			name: "update one existing column, all column value rules are mutually exclusive",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			updateColRules: []ColumnUpdate{
				{
					ColName: "A",
					RowValues: []RowValues{
						{
							RowIndexes: series.New([]bool{true, true, false, false, false}, series.Bool, ""),
							Values:     series.New([]int{2, 4, 6, 8, 10}, series.Int, ""),
						},
						{
							RowIndexes: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
							Values:     series.New([]int{3, 6, 9, 12, 15}, series.Int, ""),
						},
					},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
				},
			},
			want: New(
				series.New([]int{2, 4, -1, 12, 15}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
		},
		{
			name: "update existing one column, all column value rules are not mutually exclusive",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			updateColRules: []ColumnUpdate{
				{
					ColName: "A",
					RowValues: []RowValues{
						{
							RowIndexes: series.New([]bool{true, true, false, true, false}, series.Bool, ""),
							Values:     series.New([]int{2, 4, 6, 8, 10}, series.Int, ""),
						},
						{
							RowIndexes: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
							Values:     series.New([]int{3, 6, 9, 12, 15}, series.Int, ""),
						},
					},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
				},
			},
			want: New(
				series.New([]int{2, 4, -1, 8, 15}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
		},
		{
			name: "update multiple columns",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			updateColRules: []ColumnUpdate{
				{
					ColName: "D",
					RowValues: []RowValues{
						{
							RowIndexes: series.New([]bool{true, true, false, true, false}, series.Bool, ""),
							Values:     series.New([]int{2, 4, 6, 8, 10}, series.Int, ""),
						},
						{
							RowIndexes: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
							Values:     series.New([]int{3, 6, 9, 12, 15}, series.Int, ""),
						},
					},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
				},
				{
					ColName: "C",
					RowValues: []RowValues{
						{
							RowIndexes: series.New([]bool{true, true, true, true, false}, series.Bool, ""),
							Values:     series.New([]int{2, 4, 6, 8, 10}, series.Int, ""),
						},
						{
							RowIndexes: series.New([]bool{false, false, false, true, true}, series.Bool, ""),
							Values:     series.New([]int{3, 6, 9, 12, 15}, series.Int, ""),
						},
					},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
				},
			},
			want: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
				series.New([]int{2, 4, 6, 8, 15}, series.Int, "C"),
				series.New([]int{2, 4, -1, 8, 15}, series.Int, "D"),
			),
		},
		{
			name: "update only specified default",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			updateColRules: []ColumnUpdate{
				{
					ColName:      "C",
					RowValues:    []RowValues{},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
				},
			},
			want: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
				series.New([]int{-1, -1, -1, -1, -1}, series.Int, "C"),
			),
		},
		{
			name: "error when values in column dimension is different with table",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			updateColRules: []ColumnUpdate{
				{
					ColName: "A",
					RowValues: []RowValues{
						{
							RowIndexes: series.New([]bool{true, true, false, true, false, false}, series.Bool, ""),
							Values:     series.New([]int{2, 4, 6, 8, 10, 12}, series.Int, ""),
						},
					},
					DefaultValue: series.New([]int{-1, -1, -1, -1, -1, -1}, series.Int, ""),
				},
			},
			wantErr:    true,
			errMessage: "failed set value for column: A with err: indexing error: index dimensions mismatch",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inputTable.UpdateColumns(tt.updateColRules)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMessage)
				return
			}
			assert.Equal(t, tt.want, tt.inputTable)
		})
	}
}
