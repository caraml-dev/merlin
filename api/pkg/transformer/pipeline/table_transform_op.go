package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/scaler"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/opentracing/opentracing-go"
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
	span, _ := opentracing.StartSpanFromContext(context, "pipeline.TableTransformOp")
	defer span.Finish()

	inputTableName := t.tableTransformSpec.InputTable
	outputTableName := t.tableTransformSpec.OutputTable

	span.SetTag("table.input", inputTableName)
	span.SetTag("table.output", outputTableName)

	inputTable, err := getTable(env, inputTableName)
	if err != nil {
		return err
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
			resultTable, err = updateColumns(env, step.UpdateColumns, resultTable)
			if err != nil {
				return err
			}
		}

		if step.ScaleColumns != nil {
			columnValues := make(map[string]interface{}, len(step.ScaleColumns))
			for _, scalerSpec := range step.ScaleColumns {
				col := scalerSpec.Column
				colSeries := resultTable.Col(col)
				if err := colSeries.IsNumeric(); err != nil {
					return err
				}

				scalerImpl, err := scaler.NewScaler(scalerSpec)
				if err != nil {
					return err
				}
				scaledValues, err := scalerImpl.Scale(colSeries.GetRecords())
				if err != nil {
					return err
				}
				columnValues[col] = scaledValues
				if err := resultTable.UpdateColumnsRaw(columnValues); err != nil {
					return err
				}
			}
		}

		if step.EncodeColumns != nil {
			columnValues := make(map[string]interface{})
			for _, encodeColumn := range step.EncodeColumns {
				encoder, err := getEncoder(env, encodeColumn.Encoder)
				if err != nil {
					return err
				}
				for _, column := range encodeColumn.Columns {
					values := resultTable.Col(column).GetRecords()
					encodedValues, err := encoder.Encode(values, column)
					if err != nil {
						return err
					}
					for col, value := range encodedValues {
						columnValues[col] = value
					}
				}
			}
			if err := resultTable.UpdateColumnsRaw(columnValues); err != nil {
				return err
			}
		}

		if step.FilterRow != nil {
			rowIndexes, err := seriesFromExpression(env, step.FilterRow.Condition)
			if err != nil {
				return fmt.Errorf("error evaluating filter row expresion: %s. Err: %s", step.FilterRow, err)
			}
			if err := resultTable.FilterRow(rowIndexes); err != nil {
				return err
			}
		}

		if step.SliceRow != nil {
			sliceRow := step.SliceRow
			err := resultTable.SliceRow(int(sliceRow.Start), int(sliceRow.End))
			if err != nil {
				return err
			}
		}
	}

	env.SetSymbol(outputTableName, resultTable)
	env.LogOperation("table_transform", outputTableName)
	return nil
}

func getEncoder(env *Environment, encoderName string) (Encoder, error) {
	encoderRaw := env.symbolRegistry[encoderName]
	if encoderRaw == nil {
		return nil, fmt.Errorf("encoder %s is not declared", encoderName)
	}
	encodeImpl, ok := encoderRaw.(Encoder)
	if !ok {
		return nil, fmt.Errorf("variable %s is not encoder", encoderName)
	}
	return encodeImpl, nil
}

func updateColumns(env *Environment, specs []*spec.UpdateColumn, resultTable *table.Table) (*table.Table, error) {
	updateColumnRules := make([]table.ColumnUpdate, 0, len(specs))
	for _, updateSpec := range specs {
		columnName := updateSpec.Column
		rule := table.ColumnUpdate{
			ColName: columnName,
		}
		// if conditions is not specified, it will work like set default value
		if len(updateSpec.Conditions) == 0 {
			values, err := seriesFromExpression(env, updateSpec.Expression)
			if err != nil {
				return nil, fmt.Errorf("error evaluating expression for column %s: %v. Err: %s", columnName, updateSpec.Expression, err)
			}

			rule.DefaultValue = values
			updateColumnRules = append(updateColumnRules, rule)
			continue
		}

		columnValueRules := make([]table.RowValues, 0, len(updateSpec.Conditions))
		for _, condition := range updateSpec.Conditions {
			if condition.Default != nil {
				defaultValue, err := seriesFromExpression(env, condition.Default.Expression)
				if err != nil {
					return nil, fmt.Errorf("error evaluation default value for column %s: %v. Err: %s", columnName, condition.Default.Expression, err)
				}
				rule.DefaultValue = defaultValue
				continue
			}

			rowIndex, err := seriesFromExpression(env, condition.If)
			if err != nil {
				return nil, fmt.Errorf("error evaluating expression for condition: %s with err: %s", condition.If, err)
			}

			columnValue, err := subsetSeriesFromExpression(env, condition.Expression, rowIndex)
			if err != nil {
				return nil, fmt.Errorf("error evaluation column rule value for column %s: %v. Err: Ts", columnName, condition.Default.Expression)
			}

			colRuleValue := table.RowValues{
				RowIndexes: rowIndex,
				Values:     columnValue,
			}
			columnValueRules = append(columnValueRules, colRuleValue)
		}
		rule.RowValues = columnValueRules
		updateColumnRules = append(updateColumnRules, rule)
	}
	err := resultTable.UpdateColumns(updateColumnRules)
	if err != nil {
		return nil, err
	}
	return resultTable, nil
}
