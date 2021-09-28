package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/scaler"
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
