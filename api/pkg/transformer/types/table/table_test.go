package table

import (
	"errors"
	"testing"

	"github.com/go-gota/gota/dataframe"
	gotaSeries "github.com/go-gota/gota/series"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
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
