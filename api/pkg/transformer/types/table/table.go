package table

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"github.com/go-gota/gota/dataframe"
	gota "github.com/go-gota/gota/series"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

type Table struct {
	dataFrame *dataframe.DataFrame
}

func NewTable(df *dataframe.DataFrame) *Table {
	return &Table{dataFrame: df}
}

func New(se ...*series.Series) *Table {
	ss := make([]gota.Series, 0, 0)
	for _, gs := range se {
		ss = append(ss, *gs.Series())
	}

	df := dataframe.New(ss...)
	return &Table{dataFrame: &df}
}

func NewRaw(columnValues map[string]interface{}) (*Table, error) {
	newColumns, err := createSeries(columnValues, 1)
	if err != nil {
		return nil, err
	}

	return New(newColumns...), nil
}

// Row return a table containing only the specified row
// It's similar to GetRow, however it will panic if the specified row doesn't exists in the table
// Intended to be used as built-in function in expression
func (t *Table) Row(row int) *Table {
	result, err := t.GetRow(row)
	if err != nil {
		panic(err)
	}
	return result
}

// Col return a series containing the column specified by colName
// It's similar to GetColumn, however it will panic if the specified column doesn't exists in the table
// Intended to be used as built-in function in expression
func (t *Table) Col(colName string) *series.Series {
	result, err := t.GetColumn(colName)
	if err != nil {
		panic(err)
	}
	return result
}

// GetRow return a table containing only the specified row
func (t *Table) GetRow(row int) (*Table, error) {
	if row < 0 || row >= t.dataFrame.Nrow() {
		return nil, fmt.Errorf("invalid row number, expected: 0 <= row < %d, got: %d", t.dataFrame.Nrow(), row)
	}

	subsetDataframe := t.dataFrame.Subset(row)
	if subsetDataframe.Err != nil {
		return nil, subsetDataframe.Err
	}

	return &Table{&subsetDataframe}, nil
}

// GetColumn return a series containing the column specified by colName
func (t *Table) GetColumn(colName string) (*series.Series, error) {
	s := t.dataFrame.Col(colName)
	if s.Err != nil {
		return nil, s.Err
	}

	return series.NewSeries(&s), nil
}

// NRow return number of row in the table
func (t *Table) NRow() int {
	return t.dataFrame.Nrow()
}

// ColumnNames return slice string containing the column names
func (t *Table) ColumnNames() []string {
	return t.dataFrame.Names()
}

// Columns return slice of series containing all column values
func (t *Table) Columns() []*series.Series {
	columnNames := t.ColumnNames()
	columns := make([]*series.Series, len(columnNames))
	for idx, columnName := range columnNames {
		columns[idx], _ = t.GetColumn(columnName)
	}
	return columns
}

// DataFrame return internal representation of table
func (t *Table) DataFrame() *dataframe.DataFrame {
	return t.dataFrame
}

// Copy create a separate copy of the table
func (t *Table) Copy() *Table {
	df := t.dataFrame.Copy()
	return NewTable(&df)
}

// Concat add all column from tbl to this table with restriction that the number of row in tbl is equal with this table
func (t *Table) Concat(tbl *Table) (*Table, error) {
	if t.DataFrame().Nrow() != tbl.DataFrame().Nrow() {
		return nil, fmt.Errorf("different number of row")
	}

	leftDf := *t.DataFrame()
	rightDf := *tbl.DataFrame()

	for _, col := range rightDf.Names() {
		leftDf = leftDf.Mutate(rightDf.Col(col))
	}

	t.dataFrame = &leftDf
	return t, nil
}

// DropColumns drop all columns specified in "columns" argument
// It will return error if "columns" contains column not existing in the table
func (t *Table) DropColumns(columns []string) error {
	df := t.dataFrame.Drop(columns)
	if df.Err != nil {
		return df.Err
	}
	t.dataFrame = &df
	return nil
}

// SelectColumns perform reordering of columns and potentially drop column
// It will return error if "columns" contains column not existing in the table
func (t *Table) SelectColumns(columns []string) error {
	df := t.dataFrame.Select(columns)
	if df.Err != nil {
		return df.Err
	}
	t.dataFrame = &df
	return nil
}

// RenameColumns rename multiple column name using the mapping given by "columnMap"
// It will return error if "columnMap" contains column not existing in the table
func (t *Table) RenameColumns(columnMap map[string]string) error {
	df := t.dataFrame
	columns := df.Names()
	renamedSeries := make([]gota.Series, len(columns))

	// check all column in columnMap exists in the original table
	for colName, _ := range columnMap {
		found := false
		for _, origCol := range columns {
			if colName == origCol {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("unable to rename column: unknown column: %s", colName)
		}
	}

	// rename columns
	for idx, column := range columns {
		col := df.Col(column)
		newColName, ok := columnMap[column]
		if !ok {
			newColName = column
		}

		col.Name = newColName
		renamedSeries[idx] = col
	}

	newDf := dataframe.New(renamedSeries...)
	t.dataFrame = &newDf
	return nil
}

// Sort sort the table using rule specified in sortRules
// It will return error if "sortRules" contains column not existing in the table
func (t *Table) Sort(sortRules []*spec.SortColumnRule) error {
	df := t.dataFrame

	orders := make([]dataframe.Order, len(sortRules))
	for idx, sortRule := range sortRules {
		orders[idx] = dataframe.Order{
			Colname: sortRule.Column,
			Reverse: sortRule.Order == spec.SortOrder_DESC,
		}
	}

	newDf := df.Arrange(orders...)
	if newDf.Err != nil {
		return newDf.Err
	}
	t.dataFrame = &newDf
	return nil
}

// UpdateColumnsRaw add or update existing column with values specified in columnValues map
func (t *Table) UpdateColumnsRaw(columnValues map[string]interface{}) error {
	origColumns := t.Columns()
	updateCol := map[string]bool{} //name of col that are updates
	combinedColumns := make([]*series.Series, 0)

	columnValues = broadcastScalar(columnValues, t.NRow())

	//Check through all original columns for updates
	for _, origColumn := range origColumns {
		origColumnName := origColumn.Series().Name
		val, update := columnValues[origColumnName]

		if update {
			// Update original column
			colSeries, err := series.NewInferType(val, origColumnName)
			if err != nil {
				return err
			}
			combinedColumns = append(combinedColumns, colSeries)

			// Record down columns that are updates
			updateCol[origColumnName] = true
		} else {
			// Use original column
			combinedColumns = append(combinedColumns, origColumn)
		}
	}

	// Get all column name in colMap as slice
	colNames := make([]string, len(columnValues))

	i := 0
	for k := range columnValues {
		colNames[i] = k
		i++
	}

	sort.Strings(colNames) //to ensure predictability of appended new columns

	// Append new columns
	for _, colName := range colNames {
		_, isNotNew := updateCol[colName]
		if isNotNew {
			continue
		}

		colSeries, err := series.NewInferType(columnValues[colName], colName)
		if err != nil {
			return err
		}
		combinedColumns = append(combinedColumns, colSeries)
	}

	newT := New(combinedColumns...)
	t.dataFrame = newT.dataFrame

	return nil
}

// LeftJoin perform left join with the right table on the specified joinColumn
// Return new table containing the join result
func (t *Table) LeftJoin(right *Table, joinColumns []string) (*Table, error) {
	df := t.dataFrame.LeftJoin(*right.dataFrame, joinColumns...)
	if df.Err != nil {
		return nil, df.Err
	}

	return NewTable(&df), nil
}

// RightJoin perform right join with the right table on the specified joinColumn
// Return new table containing the join result
func (t *Table) RightJoin(right *Table, joinColumns []string) (*Table, error) {
	df := t.dataFrame.RightJoin(*right.dataFrame, joinColumns...)
	if df.Err != nil {
		return nil, df.Err
	}

	return NewTable(&df), nil
}

// InnerJoin perform inner join with the right table on the specified joinColumn
// Return new table containing the join result
func (t *Table) InnerJoin(right *Table, joinColumns []string) (*Table, error) {
	df := t.dataFrame.InnerJoin(*right.dataFrame, joinColumns...)
	if df.Err != nil {
		return nil, df.Err
	}

	return NewTable(&df), nil
}

// OuterJoin perform outer join with the right table on the specified joinColumn
// Return new table containing the join result
func (t *Table) OuterJoin(right *Table, joinColumns []string) (*Table, error) {
	df := t.dataFrame.OuterJoin(*right.dataFrame, joinColumns...)
	if df.Err != nil {
		return nil, df.Err
	}

	return NewTable(&df), nil
}

// CrossJoin perform cross join with the right table on the specified joinColumn
// Return new table containing the join result
func (t *Table) CrossJoin(right *Table) (*Table, error) {
	df := t.dataFrame.CrossJoin(*right.dataFrame)
	if df.Err != nil {
		return nil, df.Err
	}

	return NewTable(&df), nil
}

func (t *Table) String() string {
	jsonTable, _ := tableToJsonSplitFormat(t)
	jsonStr, _ := json.Marshal(jsonTable)
	return fmt.Sprintf("%v", string(jsonStr))
}

func getLength(value interface{}) int {
	colValueVal := reflect.ValueOf(value)
	switch colValueVal.Kind() {
	case reflect.Slice:
		return colValueVal.Len()
	default:
		s, ok := value.(*series.Series)
		if ok {
			return s.Series().Len()
		}
		return 1
	}
}

func broadcastScalar(colMap map[string]interface{}, length int) map[string]interface{} {
	// we don't need to broadcast if all column has length = 1
	if length == 1 {
		return colMap
	}

	for k, v := range colMap {
		val := v
		colValueVal := reflect.ValueOf(v)
		switch colValueVal.Kind() {
		case reflect.Slice:
			if colValueVal.Len() > 1 {
				continue
			}

			val = colValueVal.Index(0)
		default:
			s, ok := v.(*series.Series)
			if ok {
				if s.Series().Len() > 1 {
					continue
				}

				val = s.Get(0)
			}
		}

		values := make([]interface{}, length)
		for i := range values {
			values[i] = val
		}
		colMap[k] = values
	}

	return colMap
}

func createSeries(columnValues map[string]interface{}, maxLength int) ([]*series.Series, error) {
	// ensure all values in columnValues has length either 1 or maxLength
	for k, v := range columnValues {
		valueLength := getLength(v)
		// check that length is either 1 or maxLength
		if valueLength != 1 && maxLength != 1 && valueLength != maxLength {
			return nil, fmt.Errorf("columns %s has different dimension", k)
		}

		if valueLength > maxLength {
			maxLength = valueLength
		}
	}

	columnValues = broadcastScalar(columnValues, maxLength)

	ss := make([]*series.Series, 0)
	colNames := make([]string, len(columnValues))
	// get all column name in colMap as slice
	i := 0
	for k := range columnValues {
		colNames[i] = k
		i++
	}

	sort.Strings(colNames)
	for _, colName := range colNames {
		s, err := series.NewInferType(columnValues[colName], colName)
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}

	return ss, nil
}
