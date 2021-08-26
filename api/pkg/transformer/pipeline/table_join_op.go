package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/opentracing/opentracing-go"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/table"
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

	span.SetTag("table.left", t.tableJoinSpec.LeftTable)
	span.SetTag("table.right", t.tableJoinSpec.RightTable)
	span.SetTag("table.output", t.tableJoinSpec.OutputTable)

	leftTable, err := getTable(environment, t.tableJoinSpec.LeftTable)
	if err != nil {
		return err
	}

	rightTable, err := getTable(environment, t.tableJoinSpec.RightTable)
	if err != nil {
		return err
	}

	var resultTable *table.Table

	joinColumns := []string{t.tableJoinSpec.OnColumn}
	if len(t.tableJoinSpec.OnColumns) > 0 {
		joinColumns = t.tableJoinSpec.OnColumns
	}

	switch t.tableJoinSpec.How {
	case spec.JoinMethod_LEFT:
		err := validateJoinColumns(leftTable, rightTable, joinColumns)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.LeftJoin(rightTable, joinColumns)
		if err != nil {
			return err
		}
	case spec.JoinMethod_RIGHT:
		err := validateJoinColumns(leftTable, rightTable, joinColumns)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.RightJoin(rightTable, joinColumns)
		if err != nil {
			return err
		}
	case spec.JoinMethod_INNER:
		err := validateJoinColumns(leftTable, rightTable, joinColumns)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.InnerJoin(rightTable, joinColumns)
		if err != nil {
			return err
		}
	case spec.JoinMethod_OUTER:
		err := validateJoinColumns(leftTable, rightTable, joinColumns)
		if err != nil {
			return err
		}

		resultTable, err = leftTable.OuterJoin(rightTable, joinColumns)
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

func validateJoinColumns(leftTable *table.Table, rightTable *table.Table, joinColumns []string) error {
	var errMessages []string

	for _, joinColumn := range joinColumns {
		notFoundTables := []string{}

		if _, err := leftTable.GetColumn(joinColumn); err != nil {
			notFoundTables = append(notFoundTables, "left table")
		}

		if _, err := rightTable.GetColumn(joinColumn); err != nil {
			notFoundTables = append(notFoundTables, "right table")
		}

		if len(notFoundTables) > 0 {
			errMessages = append(errMessages, fmt.Sprintf("invalid join column: column %s does not exists in %s", joinColumn, strings.Join(notFoundTables, " and ")))
		}
	}

	if len(errMessages) > 0 {
		return fmt.Errorf("%s", strings.Join(errMessages, "\n"))
	}

	return nil
}
