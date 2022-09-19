package table

import (
	"errors"
	"net/url"
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

	_, err = table.GetColumn("col_not_exists")
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
		startIdx   *int
		endIdx     *int
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
			startIdx: toPointerInt(1),
			endIdx:   toPointerInt(3),
			want: New(
				series.New([]int{2, 3}, series.Int, "A"),
				series.New([]string{"b", "c"}, series.String, "B"),
			),
		},
		{
			name: "start and end is nil",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: nil,
			endIdx:   nil,
			want: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
		},
		{
			name: "start < end, end == table row length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: toPointerInt(1),
			endIdx:   toPointerInt(5),
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
			startIdx:   toPointerInt(1),
			endIdx:     toPointerInt(6),
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "start > end",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   toPointerInt(2),
			endIdx:     toPointerInt(0),
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "start < 0",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: toPointerInt(-1),
			endIdx:   nil,
			want: New(
				series.New([]int{5}, series.Int, "A"),
				series.New([]string{"e"}, series.String, "B"),
			),
		},
		{
			name: "end < 0",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: nil,
			endIdx:   toPointerInt(-1),
			want: New(
				series.New([]int{1, 2, 3, 4}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d"}, series.String, "B"),
			),
		},
		{
			name: "start >= 0 and end < 0",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: toPointerInt(1),
			endIdx:   toPointerInt(-2),
			want: New(
				series.New([]int{2, 3}, series.Int, "A"),
				series.New([]string{"b", "c"}, series.String, "B"),
			),
		},
		{
			name: "start < 0 and end > 0",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx: toPointerInt(-2),
			endIdx:   toPointerInt(4),
			want: New(
				series.New([]int{4}, series.Int, "A"),
				series.New([]string{"d"}, series.String, "B"),
			),
		},
		{
			name: "start > num of length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   toPointerInt(6),
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "start < -1 * num of length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   toPointerInt(-6),
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "end > num of length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   nil,
			endIdx:     toPointerInt(6),
			wantErr:    true,
			errMessage: "failed slice col: A due to: slice index out of bounds",
		},
		{
			name: "end < -1 * num of length",
			inputTable: New(
				series.New([]int{1, 2, 3, 4, 5}, series.Int, "A"),
				series.New([]string{"a", "b", "c", "d", "e"}, series.String, "B"),
			),
			startIdx:   nil,
			endIdx:     toPointerInt(-6),
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

func TestTable_RecordsFromCsv(t *testing.T) {
	tests := []struct {
		name       string
		filePath   *url.URL
		expRecords [][]string
		wantError  bool
		expError   error
	}{
		{
			name: "success: blank local file",
			filePath: &url.URL{
				Path: "testdata/blank.csv",
			},
			wantError:  false,
			expRecords: nil,
		},
		{
			name: "success: header only local file",
			filePath: &url.URL{
				Path: "testdata/header_only.csv",
			},
			wantError:  false,
			expRecords: [][]string{{"First Name", "Last Name", "Age", "Weight", "Is VIP"}},
		},
		{
			name: "success: normal local file",
			filePath: &url.URL{
				Path: "testdata/normal.csv",
			},
			wantError: false,
			expRecords: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "TRUE"},
				{"Banana", "Man", "18", "68", "FALSE"},
				{"Zara", "Vuitton", "35", "75", "TRUE"},
				{"Sandra", "Zawaska", "32", "55", "FALSE"},
				{"Merlion", "Krabby", "23", "57.22", "FALSE"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := RecordsFromCsv(tt.filePath)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expRecords, records)
		})
	}
}

func TestTable_RecordsFromParquet(t *testing.T) {
	tests := []struct {
		name       string
		filePath   *url.URL
		expRecords [][]string
		expColType map[string]gotaSeries.Type
		wantError  bool
		expError   error
	}{
		{
			name: "success: header only local file",
			filePath: &url.URL{
				Path: "testdata/header_only.parquet",
			},
			wantError:  false,
			expRecords: [][]string{{"First Name", "Last Name", "Age", "Weight", "Is VIP"}},
			expColType: map[string]gotaSeries.Type{
				"First Name": gotaSeries.Int,
				"Last Name":  gotaSeries.Int,
				"Age":        gotaSeries.Int,
				"Weight":     gotaSeries.Int,
				"Is VIP":     gotaSeries.Int,
			},
		},
		{
			name: "success: normal local file",
			filePath: &url.URL{
				Path: "testdata/normal.parquet",
			},
			wantError: false,
			expRecords: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			expColType: map[string]gotaSeries.Type{
				"First Name": gotaSeries.String,
				"Last Name":  gotaSeries.String,
				"Age":        gotaSeries.Int,
				"Weight":     gotaSeries.Float,
				"Is VIP":     gotaSeries.Bool,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, colType, err := RecordsFromParquet(tt.filePath)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expRecords, records)
			assert.Equal(t, tt.expColType, colType)
		})
	}
}

func toPointerInt(val int) *int {
	return &val
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

func TestTable_NewFromRecords(t *testing.T) {
	type args struct {
		schema []*spec.Schema
	}
	tests := []struct {
		name      string
		records   [][]string
		colType   map[string]gotaSeries.Type
		args      args
		expTable  *Table
		wantError bool
		expError  error
	}{
		{
			name:    "error: no data",
			records: [][]string{{"First Name", "Last Name", "Age", "Weight", "Is VIP"}},
			args: args{
				schema: []*spec.Schema{
					{
						Name: "First Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Last Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Age",
						Type: spec.Schema_INT,
					},
					{
						Name: "Weight",
						Type: spec.Schema_FLOAT,
					},
					{
						Name: "Is VIP",
						Type: spec.Schema_BOOL,
					},
				},
			},
			wantError: true,
			expError:  errors.New("no data found"),
		},
		{
			name: "error: Column name of schema not found in header",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			colType: map[string]gotaSeries.Type{
				"First Name": gotaSeries.String,
				"Last Name":  gotaSeries.String,
				"Age":        gotaSeries.Int,
				"Weight":     gotaSeries.Float,
				"Is VIP":     gotaSeries.Bool,
			},
			args: args{
				schema: []*spec.Schema{
					{
						Name: "First Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Last Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "age",
						Type: spec.Schema_INT,
					},
					{
						Name: "Weight",
						Type: spec.Schema_FLOAT,
					},
					{
						Name: "Is VIP",
						Type: spec.Schema_BOOL,
					},
				},
			},
			wantError: true,
			expError:  errors.New("column name of schema age not found in header of file"),
		},
		{
			name: "error: Unsupported schema type",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			args: args{
				schema: []*spec.Schema{
					{
						Name: "First Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Last Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Age",
						Type: spec.Schema_INT,
					},
					{
						Name: "Weight",
						Type: -1,
					},
					{
						Name: "Is VIP",
						Type: spec.Schema_BOOL,
					},
				},
			},
			wantError: true,
			expError:  errors.New("unsupported column type option for schema -1"),
		},
		{
			name: "success: Table with data of correct type created",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			args: args{
				schema: []*spec.Schema{
					{
						Name: "First Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Last Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Age",
						Type: spec.Schema_INT,
					},
					{
						Name: "Weight",
						Type: spec.Schema_FLOAT,
					},
					{
						Name: "Is VIP",
						Type: spec.Schema_BOOL,
					},
				},
			},
			expTable: New(
				series.New([]string{"Apple", "Banana", "Zara", "Sandra", "Merlion"}, series.String, "First Name"),
				series.New([]string{"Cider", "Man", "Vuitton", "Zawaska", "Krabby"}, series.String, "Last Name"),
				series.New([]int{25, 18, 35, 32, 23}, series.Int, "Age"),
				series.New([]float64{48.8, 68, 75, 55, 57.22}, series.Float, "Weight"),
				series.New([]bool{true, false, true, false, false}, series.Bool, "Is VIP"),
			),
			wantError: false,
		},
		{
			name: "success: Table with data with auto-detect type (incomplete schema)",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			args: args{
				schema: []*spec.Schema{
					{
						Name: "First Name",
						Type: spec.Schema_STRING,
					},
					{
						Name: "Age",
						Type: spec.Schema_INT,
					},
					{
						Name: "Is VIP",
						Type: spec.Schema_BOOL,
					},
				},
			},
			expTable: New(
				series.New([]string{"Apple", "Banana", "Zara", "Sandra", "Merlion"}, series.String, "First Name"),
				series.New([]string{"Cider", "Man", "Vuitton", "Zawaska", "Krabby"}, series.String, "Last Name"),
				series.New([]int{25, 18, 35, 32, 23}, series.Int, "Age"),
				series.New([]float64{48.8, 68, 75, 55, 57.22}, series.Float, "Weight"),
				series.New([]bool{true, false, true, false, false}, series.Bool, "Is VIP"),
			),
			wantError: false,
		},
		{
			name: "success: no schema, colType",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			colType: map[string]gotaSeries.Type{
				"First Name": gotaSeries.String,
				"Last Name":  gotaSeries.String,
				"Age":        gotaSeries.Int,
				"Weight":     gotaSeries.Float,
				"Is VIP":     gotaSeries.String,
			},
			args: args{
				schema: nil,
			},
			expTable: New(
				series.New([]string{"Apple", "Banana", "Zara", "Sandra", "Merlion"}, series.String, "First Name"),
				series.New([]string{"Cider", "Man", "Vuitton", "Zawaska", "Krabby"}, series.String, "Last Name"),
				series.New([]int{25, 18, 35, 32, 23}, series.Int, "Age"),
				series.New([]float64{48.8, 68, 75, 55, 57.22}, series.Float, "Weight"),
				series.New([]string{"true", "false", "true", "false", "false"}, series.String, "Is VIP"),
			),
			wantError: false,
		},
		{
			name: "success: no schema, no colType",
			records: [][]string{
				{"First Name", "Last Name", "Age", "Weight", "Is VIP"},
				{"Apple", "Cider", "25", "48.8", "true"},
				{"Banana", "Man", "18", "68", "false"},
				{"Zara", "Vuitton", "35", "75", "true"},
				{"Sandra", "Zawaska", "32", "55", "false"},
				{"Merlion", "Krabby", "23", "57.22", "false"},
			},
			args: args{
				schema: nil,
			},
			expTable: New(
				series.New([]string{"Apple", "Banana", "Zara", "Sandra", "Merlion"}, series.String, "First Name"),
				series.New([]string{"Cider", "Man", "Vuitton", "Zawaska", "Krabby"}, series.String, "Last Name"),
				series.New([]int{25, 18, 35, 32, 23}, series.Int, "Age"),
				series.New([]float64{48.8, 68, 75, 55, 57.22}, series.Float, "Weight"),
				series.New([]bool{true, false, true, false, false}, series.Bool, "Is VIP"),
			),
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileTable, err := NewFromRecords(tt.records, tt.colType, tt.args.schema)
			if tt.wantError {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expTable, fileTable)
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
						{
							RowIndexes: series.New([]bool{false, false, true, false, false}, series.Bool, ""),
							Values:     series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
						},
					},
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
						{
							RowIndexes: series.New([]bool{false, false, true, false, false}, series.Bool, ""),
							Values:     series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
						},
					},
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
						{
							RowIndexes: series.New([]bool{false, false, true, false, false}, series.Bool, ""),
							Values:     series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
						},
					},
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
						{
							RowIndexes: series.New([]bool{false, false, false, false, false}, series.Bool, ""),
							Values:     series.New([]int{-1, -1, -1, -1, -1}, series.Int, ""),
						},
					},
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
						{
							RowIndexes: series.New([]bool{false, false, true, false, true, true}, series.Bool, ""),
							Values:     series.New([]int{-1, -1, -1, -1, -1, -1}, series.Int, ""),
						},
					},
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
