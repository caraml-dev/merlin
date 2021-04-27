package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type TableTransformOp struct {
	tableTransformSpec *spec.TableTransformation
}

func NewTableTransformOp(tableTransformSpec *spec.TableTransformation) Op {
	return &TableTransformOp{
		tableTransformSpec: tableTransformSpec,
	}
}

func (t TableTransformOp) Execute(context context.Context, env *Environment) error {
	inputTableName := t.tableTransformSpec.InputTable
	outputTableName := t.tableTransformSpec.OutputTable

	inputTableRaw := env.symbolRegistry[inputTableName]
	if inputTableRaw == nil {
		return fmt.Errorf("table %s is not declared", inputTableName)
	}

	inputTable, ok := inputTableRaw.(*table.Table)
	if !ok {
		return fmt.Errorf("variable %s is not a table", inputTableName)
	}

	resultTable := inputTable.Copy()
	for _, step := range t.tableTransformSpec.Steps {
		if step.DropColumns != nil {
			err := resultTable.DropColumns(step.DropColumns)
			if err != nil {
				return err
			}
		}

		if step.SelectColumns != nil {
			err := resultTable.SelectColumns(step.SelectColumns)
			if err != nil {
				return err
			}
		}

		if step.RenameColumns != nil {
			err := resultTable.RenameColumns(step.RenameColumns)
			if err != nil {
				return err
			}
		}

		if step.Sort != nil {
			err := resultTable.Sort(step.Sort)
			if err != nil {
				return err
			}
		}

		if step.UpdateColumns != nil {
			columnValues := make(map[string]interface{}, len(step.UpdateColumns))
			for _, updateSpec := range step.UpdateColumns {
				columnName := updateSpec.Column
				result, err := evalExpression(env, updateSpec.Expression)
				if err != nil {
					return fmt.Errorf("error evaluating expression for column %s: %v", columnName, updateSpec.Expression)
				}
				columnValues[columnName] = result
			}

			err := resultTable.UpdateColumnsRaw(columnValues)
			if err != nil {
				return err
			}
		}
	}

	env.SetSymbol(outputTableName, resultTable)
	return nil
}
