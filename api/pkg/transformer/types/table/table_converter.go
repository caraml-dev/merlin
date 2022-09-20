package table

import (
	"errors"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/gojek/merlin/pkg/transformer/types/series"
)

const (
	columnsJsonKey = "columns"
	dataJsonKey    = "data"
	// RowIDColumn is reserved column name in standard transformer when UPI_V1 protocol applied
	RowIDColumn = "row_id"
)

// TableToJson converts table into JSON with defined format
func TableToJson(tbl *Table, format spec.FromTable_JsonFormat) (interface{}, error) {
	switch format {
	case spec.FromTable_RECORD:
		return tableToJsonRecordFormat(tbl)
	case spec.FromTable_VALUES:
		return tableToJsonValuesFormat(tbl)
	case spec.FromTable_SPLIT:
		return tableToJsonSplitFormat(tbl)
	default:
		return nil, errors.New("unsupported format")
	}
}

func getRowIDValues(tbl *Table) []string {
	rowIDSeries, _ := tbl.GetColumn(RowIDColumn)
	var rowIDValues []string
	if rowIDSeries != nil {
		rowIDValues = rowIDSeries.Series().Records()
	}
	return rowIDValues
}

func convertToUPIColumns(cols []*series.Series) ([]*upiv1.Column, error) {
	upiCols := make([]*upiv1.Column, len(cols))
	for idx, col := range cols {
		colName := col.Series().Name
		upiColType := upiv1.Type_TYPE_STRING
		switch col.Type() {
		case series.Int:
			upiColType = upiv1.Type_TYPE_INTEGER
		case series.Float:
			upiColType = upiv1.Type_TYPE_DOUBLE
		case series.String:
			upiColType = upiv1.Type_TYPE_STRING
		default:
			return nil, fmt.Errorf("type %v is not supported in UPI", col.Type())
		}

		upiCols[idx] = &upiv1.Column{
			Name: colName,
			Type: upiColType,
		}
	}
	return upiCols, nil
}

// ToUPITable converts table into upi table
// will exclude row_id columns from upi columns
func ToUPITable(tbl *Table, name string) (*upiv1.Table, error) {
	cols := tbl.ColumnsWithExcluding([]string{RowIDColumn})
	upiCols, err := convertToUPIColumns(cols)
	if err != nil {
		return nil, err
	}

	rowIDValues := getRowIDValues(tbl)
	colValues, err := getTableRecordsForColumns(cols)
	if err != nil {
		return nil, err
	}

	upiRows := make([]*upiv1.Row, tbl.NRow())
	rowIDColExist := tbl.NRow() == len(rowIDValues)
	for rowIdx := 0; rowIdx < tbl.NRow(); rowIdx++ {
		vals := make([]*upiv1.Value, len(cols))
		for colIdx, col := range upiCols {
			colValues := colValues[col.Name]
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

// FromUPITable convert UPI table into standard transformer table
func FromUPITable(tbl *upiv1.Table) (*Table, error) {
	cols := tbl.Columns
	lastColsIdx := len(cols)
	colTypeLookup := make(map[int]upiv1.Type)
	columnNames := make([]string, lastColsIdx+1)
	for i, col := range cols {
		colTypeLookup[i] = col.Type
		columnNames[i] = col.Name
	}
	// adding row_id column
	colTypeLookup[lastColsIdx] = upiv1.Type_TYPE_STRING
	columnNames[lastColsIdx] = RowIDColumn

	getValFn := func(value *upiv1.Value, cType upiv1.Type) (any, error) {
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
			val, err := getValFn(val, colType)
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

func tableToJsonRecordFormat(tbl *Table) (interface{}, error) {
	columnsValues, err := getTableRecordsAllColumns(tbl)
	if err != nil {
		return nil, err
	}

	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()

	records := make([]interface{}, 0, dataFrame.Nrow())
	for i := 0; i < dataFrame.Nrow(); i++ {
		rowJson := make(map[string]interface{}, len(columns))
		for _, col := range columns {
			values := columnsValues[col]
			rowJson[col] = values[i]
		}
		records = append(records, rowJson)
	}
	return records, nil
}

func tableToJsonValuesFormat(tbl *Table) (interface{}, error) {
	columnsValues, err := getTableRecordsAllColumns(tbl)
	if err != nil {
		return nil, err
	}

	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()

	var records []interface{}
	for i := 0; i < dataFrame.Nrow(); i++ {
		var rowRecord []interface{}
		for _, col := range columns {
			values := columnsValues[col]
			rowRecord = append(rowRecord, values[i])
		}
		records = append(records, rowRecord)
	}
	return records, nil
}

func tableToJsonSplitFormat(tbl *Table) (interface{}, error) {
	dataFrame := tbl.DataFrame()
	columns := dataFrame.Names()
	records := make(map[string]interface{})
	records[columnsJsonKey] = columns

	values, err := tableToJsonValuesFormat(tbl)
	if err != nil {
		return nil, err
	}
	records[dataJsonKey] = values

	return records, nil
}

func getTableRecordsForColumns(columns []*series.Series) (map[string][]interface{}, error) {
	columnsValues := make(map[string][]interface{})
	for _, col := range columns {
		columnsValues[col.Series().Name] = col.GetRecords()
	}
	return columnsValues, nil
}

func getTableRecordsAllColumns(tbl *Table) (map[string][]interface{}, error) {
	columns := tbl.DataFrame().Names()

	columnsValues := make(map[string][]interface{})
	for _, col := range columns {
		valueSeries := tbl.Col(col)
		columnsValues[col] = valueSeries.GetRecords()
	}
	return columnsValues, nil
}
