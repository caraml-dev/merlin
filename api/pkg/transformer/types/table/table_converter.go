package table

import (
	"errors"

	"github.com/gojek/merlin/pkg/transformer/spec"
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
