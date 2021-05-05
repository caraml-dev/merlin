package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-gota/gota/dataframe"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type CreateTableOp struct {
	tableSpecs []*spec.Table
}

func NewCreateTableOp(tableSpecs []*spec.Table) Op {
	return &CreateTableOp{tableSpecs: tableSpecs}
}

func (c CreateTableOp) Execute(_ context.Context, env *Environment) error {
	for _, tableSpec := range c.tableSpecs {
		var t *table.Table
		if tableSpec.BaseTable != nil {
			tbl, err := createBaseTable(env, tableSpec.BaseTable)
			if err != nil {
				return fmt.Errorf("unable to create base table for %s: %w", tableSpec.Name, err)
			}
			t = tbl
		}

		if tableSpec.Columns != nil {
			tbl, err := overrideColumns(env, t, tableSpec.Columns)
			if err != nil {
				return fmt.Errorf("unable to override column for table %s: %w", tableSpec.Name, err)
			}
			t = tbl
		}

		// register to environment
		env.SetSymbol(tableSpec.Name, t)
		env.LogOperation("create_table", tableSpec.Name)
	}

	return nil
}

func createBaseTable(env *Environment, baseTableSpec *spec.BaseTable) (*table.Table, error) {
	switch baseTable := baseTableSpec.BaseTable.(type) {
	case *spec.BaseTable_FromJson:
		jsonObj, err := evalJSONPath(env, baseTable.FromJson.JsonPath)
		if err != nil {
			return nil, err
		}

		maps, err := toMaps(jsonObj)
		if err != nil {
			return nil, fmt.Errorf("invalid json pointed by %s: %w", baseTable.FromJson.JsonPath, err)
		}

		if baseTable.FromJson.AddRowNumber {
			for i, m := range maps {
				m["row_number"] = i
			}
		}

		df := dataframe.LoadMaps(maps)
		if df.Err != nil {
			return nil, df.Err
		}
		return table.NewTable(&df), nil
	case *spec.BaseTable_FromTable:
		s := env.SymbolRegistry()[baseTable.FromTable.TableName]
		if s == nil {
			return nil, fmt.Errorf("table %s is not found", baseTable.FromTable.TableName)
		}

		t, ok := s.(*table.Table)
		if !ok {
			return nil, fmt.Errorf("variable %s is not a table", baseTable.FromTable.TableName)
		}

		return t.Copy(), nil
	default:
		return nil, fmt.Errorf("unsupported base table spec %T", baseTable)
	}
}

func overrideColumns(env *Environment, t *table.Table, columns []*spec.Column) (*table.Table, error) {
	columnValues := make(map[string]interface{}, len(columns))
	for _, col := range columns {
		switch c := col.GetColumnValue().(type) {
		case *spec.Column_FromJson:
			result, err := evalJSONPath(env, c.FromJson.JsonPath)
			if err != nil {
				return nil, err
			}

			columnValues[col.Name] = result
		case *spec.Column_Expression:
			result, err := evalExpression(env, c.Expression)
			if err != nil {
				return nil, err
			}

			columnValues[col.Name] = result
		}
	}

	// create a new table
	if t == nil {
		return table.NewRaw(columnValues)
	}

	// update existing table
	err := t.UpdateColumnsRaw(columnValues)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func toMaps(jsonObj interface{}) ([]map[string]interface{}, error) {
	jsonArray, ok := jsonObj.([]interface{})
	if !ok {
		return nil, errors.New("not an array")
	}

	var result []map[string]interface{}
	for _, j := range jsonArray {
		node, ok := j.(map[string]interface{})
		if !ok {
			return nil, errors.New("not an array of JSON object")
		}
		result = append(result, node)
	}

	return result, nil
}
