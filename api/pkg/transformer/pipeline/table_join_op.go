package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/table"
	"github.com/opentracing/opentracing-go"
)

type TableJoinOp struct {
	tableJoinSpec *spec.TableJoin
}

func NewTableJoinOp(tableJoinSpec *spec.TableJoin) Op {
	return &TableJoinOp{
		tableJoinSpec: tableJoinSpec,
	}
}

func (t TableJoinOp) Execute(context context.Context, environment *Environment) error {
	span, _ := opentracing.StartSpanFromContext(context, "pipeline.TableJoin")
	defer span.Finish()

	leftTable, err := getTable(environment, t.tableJoinSpec.LeftTable)
	if err != nil {
		return err
	}

	rightTable, err := getTable(environment, t.tableJoinSpec.RightTable)
	if err != nil {
		return err
	}

	var resultTable *table.Table
	joinColumn := t.tableJoinSpec.OnColumn
	switch t.tableJoinSpec.How {
	case spec.JoinMethod_LEFT:
		err := validateJoinColumn(leftTable, rightTable, joinColumn)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.LeftJoin(rightTable, joinColumn)
		if err != nil {
			return err
		}
	case spec.JoinMethod_RIGHT:
		err := validateJoinColumn(leftTable, rightTable, joinColumn)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.RightJoin(rightTable, joinColumn)
		if err != nil {
			return err
		}
	case spec.JoinMethod_INNER:
		err := validateJoinColumn(leftTable, rightTable, joinColumn)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.InnerJoin(rightTable, joinColumn)
		if err != nil {
			return err
		}
	case spec.JoinMethod_OUTER:
		err := validateJoinColumn(leftTable, rightTable, joinColumn)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.OuterJoin(rightTable, joinColumn)
		if err != nil {
			return err
		}
	case spec.JoinMethod_CROSS:
		resultTable, err = leftTable.CrossJoin(rightTable)
		if err != nil {
			return err
		}
	case spec.JoinMethod_CONCAT:
		resultTable, err = leftTable.Concat(rightTable)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown join method: %s", t.tableJoinSpec.How)
	}

	environment.SetSymbol(t.tableJoinSpec.OutputTable, resultTable)
	environment.LogOperation("table_join", t.tableJoinSpec.OutputTable)
	return nil
}

func getTable(env *Environment, tableName string) (*table.Table, error) {
	tableRaw := env.symbolRegistry[tableName]
	if tableRaw == nil {
		return nil, fmt.Errorf("table %s is not declared", tableName)
	}

	tableOut, ok := tableRaw.(*table.Table)
	if !ok {
		return nil, fmt.Errorf("variable %s is not a table", tableName)
	}
	return tableOut, nil
}

func validateJoinColumn(leftTable *table.Table, rightTable *table.Table, joinColumn string) error {
	_, err := leftTable.GetColumn(joinColumn)
	if err != nil {
		return fmt.Errorf("invalid join column: column %s does not exists in left table", joinColumn)
	}

	_, err = rightTable.GetColumn(joinColumn)
	if err != nil {
		return fmt.Errorf("invalid join column: column %s does not exists in right table", joinColumn)
	}
	return err
}
