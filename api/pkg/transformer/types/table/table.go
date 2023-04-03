package table

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"sort"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types/converter"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/go-gota/gota/dataframe"
	gota "github.com/go-gota/gota/series"

	"github.com/bboughton/gcp-helpers/gsutil"
)

const (
	gcsScheme = "gs"
)

type Table struct {
	dataFrame *dataframe.DataFrame
}

func NewTable(df *dataframe.DataFrame) *Table {
	return &Table{dataFrame: df}
}

func New(se ...*series.Series) *Table {
	ss := make([]gota.Series, 0)
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

// NewFromUPITable convert UPI table into standard transformer table
func NewFromUPITable(tbl *upiv1.Table) (*Table, error) {
	cols := tbl.Columns
	lastColsIdx := len(cols)
	colTypeLookup := make(map[int]upiv1.Type)

	// adding +1 to add `row_id column`
	columnNames := make([]string, lastColsIdx+1)
	for i, col := range cols {
		colTypeLookup[i] = col.Type
		columnNames[i] = col.Name
	}
	// adding row_id column
	colTypeLookup[lastColsIdx] = upiv1.Type_TYPE_STRING
	columnNames[lastColsIdx] = RowIDColumn

	seriesVals := make([][]any, len(columnNames))
	for idx, row := range tbl.Rows {
		if lastColsIdx != len(row.Values) {
			return nil, fmt.Errorf("length column in a row: %d doesn't match with defined columns length %d", len(row.Values), len(cols))
		}
		for colIdx, val := range row.Values {
			colType := colTypeLookup[colIdx]
			currSeriesVal := seriesVals[colIdx]
			if currSeriesVal == nil {
				currSeriesVal = make([]any, len(tbl.Rows))
			}
			val, err := getUPIValue(val, colType)
			if err != nil {
				return nil, err
			}
			currSeriesVal[idx] = val
			seriesVals[colIdx] = currSeriesVal
		}
		// adding row_id series
		rowIdSeriesVal := seriesVals[lastColsIdx]
		if rowIdSeriesVal == nil {
			rowIdSeriesVal = make([]any, len(tbl.Rows))
		}
		rowIdSeriesVal[idx] = row.RowId
		seriesVals[len(cols)] = rowIdSeriesVal
	}
	generatedSeries := make([]*series.Series, len(seriesVals))
	for i, vals := range seriesVals {
		colType := colTypeLookup[i]
		seriesType := series.String
		if colType == upiv1.Type_TYPE_INTEGER {
			seriesType = series.Int
		} else if colType == upiv1.Type_TYPE_DOUBLE {
			seriesType = series.Float
		}
		generatedSeries[i] = series.New(vals, seriesType, columnNames[i])
	}
	return New(generatedSeries...), nil
}

func getUPIValue(value *upiv1.Value, cType upiv1.Type) (any, error) {
	if value.IsNull {
		return nil, nil
	}
	switch cType {
	case upiv1.Type_TYPE_INTEGER:
		return value.IntegerValue, nil
	case upiv1.Type_TYPE_STRING:
		return value.StringValue, nil
	case upiv1.Type_TYPE_DOUBLE:
		return value.DoubleValue, nil
	default:
		return nil, fmt.Errorf("got unexpected type %v", cType)
	}
}

// RecordsFromParquet reads a parquet file which may be local (uploaded together with model) or global (from gcs).
// The root directory for local will be same as where the model is stored (artifacts folder).
// Data will be returned as [][]string where row 0 is the header containing all the column names.
// The types for each column, defined in the parquet file will also be returned as a map
func RecordsFromParquet(filePath *url.URL) ([][]string, map[string]gota.Type, error) {
	var r io.ReadSeeker

	if filePath.Scheme == gcsScheme {
		// Global file (GCS)
		ctx := context.Background()
		data, err := gsutil.ReadFile(ctx, filePath.String())

		if err != nil {
			return nil, nil, err
		}

		r = bytes.NewReader(data)
	} else {
		// Local file
		file, err := os.Open(filePath.String())
		if err != nil {
			return nil, nil, err
		}
		r = file
		defer file.Close() //nolint:errcheck
	}

	fr, err := goparquet.NewFileReader(r)
	if err != nil {
		return nil, nil, err
	}

	// Get header & schema
	schema := fr.GetSchemaDefinition()
	header := make([]string, 0)
	colType := make(map[string]gota.Type, 0)
	for _, schemaCol := range schema.RootColumn.Children {
		colName := schemaCol.SchemaElement.GetName()
		header = append(header, colName)
		switch schemaCol.SchemaElement.GetType() {
		case parquet.Type_BOOLEAN:
			colType[colName] = gota.Bool
		case parquet.Type_INT32, parquet.Type_INT64, parquet.Type_INT96:
			colType[colName] = gota.Int
		case parquet.Type_DOUBLE, parquet.Type_FLOAT:
			colType[colName] = gota.Float
		case parquet.Type_BYTE_ARRAY, parquet.Type_FIXED_LEN_BYTE_ARRAY:
			colType[colName] = gota.String
		}

	}

	records := make([][]string, 0)
	records = append(records, header) // column name in first row

	// map col name to col number
	colMap := make(map[string]int)
	for i, colName := range header {
		colMap[colName] = i
	}

	// Build data
	for i := 0; i < int(fr.NumRows()); i++ {
		row, err := fr.NextRow()
		if err != nil {
			return nil, nil, err
		}

		newRow := make([]string, len(header))
		for k, v := range row {
			if vv, ok := v.([]byte); ok {
				v = string(vv)
			}

			newRow[colMap[k]] = fmt.Sprint(v)
		}
		records = append(records, newRow)
	}

	return records, colType, nil
}

// RecordsFromCsv reads a csv file which may be local (uploaded together with model) or global (from gcs).
// The root directory for local will be same as where the model is stored (artifacts folder).
// Data will be returned as [][]string where row 0 is the header containing all the column names.
func RecordsFromCsv(filePath *url.URL) ([][]string, error) {
	var r io.Reader

	if filePath.Scheme == gcsScheme {
		// Global file (GCS)
		ctx := context.Background()
		data, err := gsutil.ReadFile(ctx, filePath.String())

		if err != nil {
			return nil, err
		}

		r = bytes.NewReader(data)
	} else {
		// Local file
		file, err := os.Open(filePath.String())
		if err != nil {
			return nil, err
		}
		r = file
		defer file.Close() //nolint:errcheck
	}

	csvReader := csv.NewReader(r)
	records, err := csvReader.ReadAll()

	if err != nil {
		return nil, err
	}

	return records, nil
}

// NewFromRecords create a new table from records in the form [][]string
// The first row of the records array shall contain all the column names.
// This function will check that there are at least 2 rows in records (header and data).
// It will also auto-detect the data types if schema of the column is not provided.
// before creating a new table to be returned
// records may be supplied with its column types as a map of column name to gota types.
// Note: Schema is user specified, colType is type defined in files such as parquet. It may be set to nil.
func NewFromRecords(records [][]string, colType map[string]gota.Type, schema []*spec.Schema) (*Table, error) {
	var dfSchema map[string]gota.Type

	// check table has at least 1 row of data
	if len(records) <= 1 {
		return nil, fmt.Errorf("no data found")
	}
	// create a header list for checking
	header := make(map[string]bool)
	for _, colName := range records[0] {
		header[colName] = true
	}

	// prepare schema for dataframe
	if colType == nil {
		dfSchema = make(map[string]gota.Type)
	} else {
		dfSchema = colType
	}
	for _, colSpec := range schema {
		// validate colname defined in schema exist
		if !header[colSpec.GetName()] {
			return nil, fmt.Errorf("column name of schema %s not found in header of file", colSpec.GetName())
		}

		// build the schema map for dataframe
		switch colType := colSpec.GetType(); colType {
		case spec.Schema_STRING:
			dfSchema[colSpec.GetName()] = gota.String
		case spec.Schema_INT:
			dfSchema[colSpec.GetName()] = gota.Int
		case spec.Schema_FLOAT:
			dfSchema[colSpec.GetName()] = gota.Float
		case spec.Schema_BOOL:
			dfSchema[colSpec.GetName()] = gota.Bool
		default:
			return nil, fmt.Errorf("unsupported column type option for schema %s", colType)
		}
	}

	// build dataframe
	df := dataframe.LoadRecords(records,
		dataframe.DetectTypes(true),
		dataframe.WithTypes(dfSchema),
	)

	return NewTable(&df), nil
}

// ToUPITable converts table into upi table
// will exclude row_id columns from upi columns
func (t *Table) ToUPITable(name string) (*upiv1.Table, error) {
	cols := t.ColumnsExcluding([]string{RowIDColumn})
	upiCols, err := series.ConvertToUPIColumns(cols)
	if err != nil {
		return nil, err
	}

	rowIDValues := getRowIDValues(t)
	allColValues, err := getTableRecordsForColumns(cols)
	if err != nil {
		return nil, err
	}

	upiRows := make([]*upiv1.Row, t.NRow())
	rowIDColExist := t.NRow() == len(rowIDValues)
	for rowIdx := 0; rowIdx < t.NRow(); rowIdx++ {
		vals := make([]*upiv1.Value, len(cols))
		for colIdx, col := range upiCols {
			colValues := allColValues[col.Name]
			colVal := colValues[rowIdx]
			val := &upiv1.Value{}
			if colVal == nil {
				val.IsNull = true
				vals[colIdx] = val
				continue
			}

			switch col.Type {
			case upiv1.Type_TYPE_INTEGER:
				intVal, err := converter.ToInt64(colVal)
				if err != nil {
					return nil, err
				}
				val.IntegerValue = intVal
			case upiv1.Type_TYPE_DOUBLE:
				floatVal, err := converter.ToFloat64(colVal)
				if err != nil {
					return nil, err
				}
				val.DoubleValue = floatVal
			case upiv1.Type_TYPE_STRING:
				strVal, err := converter.ToString(colVal)
				if err != nil {
					return nil, err
				}
				val.StringValue = strVal
			default:
				return nil, fmt.Errorf("not supported type")
			}
			vals[colIdx] = val
		}
		var rowID string
		if rowIDColExist {
			rowID = rowIDValues[rowIdx]
		}
		rowVal := &upiv1.Row{
			RowId:  rowID,
			Values: vals,
		}
		upiRows[rowIdx] = rowVal
	}

	upiTbl := &upiv1.Table{
		Name:    name,
		Columns: upiCols,
		Rows:    upiRows,
	}
	return upiTbl, nil
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

func (t *Table) ColumnExist(columnName string) bool {
	columnNames := t.ColumnNames()
	for _, colName := range columnNames {
		if colName == columnName {
			return true
		}
	}
	return false
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

// ColumnsExcluding retrieve all the columns values except given excluding columns
func (t *Table) ColumnsExcluding(excludingCols []string) []*series.Series {
	columnNames := t.ColumnNames()
	columns := make([]*series.Series, 0)
	exludingColLookup := make(map[string]bool)
	for _, col := range excludingCols {
		exludingColLookup[col] = true
	}
	for _, colName := range columnNames {
		if exludingColLookup[colName] {
			continue
		}

		col, _ := t.GetColumn(colName)
		columns = append(columns, col)
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
	for colName := range columnMap {
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
	updateCol := map[string]bool{} // name of col that are updates
	combinedColumns := make([]*series.Series, 0)

	columnValues = broadcastScalar(columnValues, t.NRow())

	// Check through all original columns for updates
	for _, origColumn := range origColumns {
		origColumnName := origColumn.Series().Name
		val, update := columnValues[origColumnName]

		updatedColumn := origColumn
		if update {
			// Update original column
			colSeries, err := series.NewInferType(val, origColumnName)
			if err != nil {
				return err
			}
			updatedColumn = colSeries

			// Record down columns that are updates
			updateCol[origColumnName] = true
		}

		// Add column, either the origColumn or the updated one
		combinedColumns = append(combinedColumns, updatedColumn)
	}

	// Get all column name in colMap as slice
	colNames := make([]string, len(columnValues))

	i := 0
	for k := range columnValues {
		colNames[i] = k
		i++
	}

	sort.Strings(colNames) // to ensure predictability of appended new columns

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

// SliceRow will slice all columns in a table based on supplied start and end index
func (t *Table) SliceRow(start, end *int) error {
	df := t.dataFrame
	startIdx, endIdx := series.NormalizeIndex(start, end, t.NRow())
	newDf := df.Slice(startIdx, endIdx)
	if newDf.Err != nil {
		return newDf.Err
	}
	t.dataFrame = &newDf
	return nil
}

// FilterRow will select subset of table based on the supplied indexes
func (t *Table) FilterRow(rowIndexes *series.Series) error {
	df := t.dataFrame
	var idxs interface{}
	if rowIndexes != nil {
		idxs = *(rowIndexes.Series())
	}
	newDf := df.Subset(idxs)
	if newDf.Err != nil {
		return newDf.Err
	}
	t.dataFrame = &newDf
	return nil
}

// RowValues indicate which values that needs to be set in a set of row
type RowValues struct {
	RowIndexes *series.Series
	Values     *series.Series
}

// ColumnUpdate is a rule to update a column, this contains of name of column and `RowValues` indicate values for set of row in a column
// also have defaultValue
type ColumnUpdate struct {
	RowValues []RowValues
	ColName   string
}

// UpdateColumns is method to update multiple columns given list of rules for update column (ColumnUpdate)
func (t *Table) UpdateColumns(columnUpdates []ColumnUpdate) error {
	df := t.dataFrame
	columnUpdateRules := make([]dataframe.ColumnUpdate, 0, len(columnUpdates))

	sort.Slice(columnUpdates, func(i, j int) bool {
		return columnUpdates[i].ColName < columnUpdates[j].ColName
	})

	for _, updateRule := range columnUpdates {
		rowValues := make([]dataframe.RowValues, 0, len(updateRule.RowValues))
		for _, colValueRule := range updateRule.RowValues {

			idx := colValueRule.RowIndexes
			if idx == nil {
				continue
			}

			rowValues = append(rowValues, dataframe.RowValues{
				Values:     *colValueRule.Values.Series(),
				RowIndexes: *(idx.Series()),
			})
		}

		columnUpdateRule := dataframe.ColumnUpdate{
			ColName:   updateRule.ColName,
			RowValues: rowValues,
		}

		columnUpdateRules = append(columnUpdateRules, columnUpdateRule)
	}

	newDf := df.UpdateColumns(columnUpdateRules)
	if newDf.Err != nil {
		return newDf.Err
	}
	t.dataFrame = &newDf
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

			val = colValueVal.Index(0).Interface()
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
