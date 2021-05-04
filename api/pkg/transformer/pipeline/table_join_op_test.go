package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/merlin/pkg/transformer/jsonpath"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/expression"
	"github.com/gojek/merlin/pkg/transformer/types/series"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

func TestTableJoinOp_Execute(t *testing.T) {
	compiledExpression := expression.NewStorage()
	compiledJsonPath := jsonpath.NewStorage()

	customerTable := table.New(
		series.New([]interface{}{1, 2, 3, 4}, series.Int, "customer_id"),
		series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Elise"}, series.String, "name"),
		series.New([]interface{}{24, 25, 23, 20}, series.Int, "age"),
	)

	orderTable := table.New(
		series.New([]interface{}{111, 222, 333, 444}, series.Int, "oid"),
		series.New([]interface{}{3, 3, 2, 4}, series.Int, "customer_id"),
		series.New([]interface{}{500, 10, 12, 24}, series.Int, "amount"),
	)

	tableA := table.New(
		series.New([]interface{}{1, 2, 3}, series.Int, "col_a"),
	)

	tableB := table.New(
		series.New([]interface{}{11, 22, 33}, series.Int, "col_b"),
	)

	env := NewEnvironment(&CompiledPipeline{
		compiledJsonpath:   compiledJsonPath,
		compiledExpression: compiledExpression,
	})

	env.SetSymbol("customer_table", customerTable)
	env.SetSymbol("order_table", orderTable)
	env.SetSymbol("table_a", tableA)
	env.SetSymbol("table_b", tableB)
	env.SetSymbol("integer_var", 12345)

	tests := []struct {
		name         string
		joinSpec     *spec.TableJoin
		env          *Environment
		expVariables map[string]interface{}
		wantErr      bool
		expError     error
	}{
		{
			name: "success: left join two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				On:          "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "oid"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount")),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: right join two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "order_table",
				RightTable:  "customer_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_RIGHT,
				On:          "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4, 1}, series.Int, "customer_id"),
					series.New([]interface{}{333, 111, 222, 444, nil}, series.Int, "oid"),
					series.New([]interface{}{12, 500, 10, 24, nil}, series.Int, "amount"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise", "Ramesh"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20, 24}, series.Int, "age"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: inner join two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_INNER,
				On:          "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{333, 111, 222, 444}, series.Int, "oid"),
					series.New([]interface{}{12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: outer join two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_OUTER,
				On:          "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "oid"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: cross join two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_a",
				RightTable:  "table_b",
				OutputTable: "result_table",
				How:         spec.JoinMethod_CROSS,
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 1, 1, 2, 2, 2, 3, 3, 3}, series.Int, "col_a"),
					series.New([]interface{}{11, 22, 33, 11, 22, 33, 11, 22, 33}, series.Int, "col_b"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: concat  two table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_a",
				RightTable:  "table_b",
				OutputTable: "result_table",
				How:         spec.JoinMethod_CONCAT,
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3}, series.Int, "col_a"),
					series.New([]interface{}{11, 22, 33}, series.Int, "col_b"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "error: join column does not exist in left table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				On:          "oid",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid join column: column oid does not exists in left table"),
		},
		{
			name: "error: join column does not exist in right table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				On:          "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid join column: column age does not exists in right table"),
		},
		{
			name: "error: not a table",
			joinSpec: &spec.TableJoin{
				LeftTable:   "integer_var",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				On:          "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("variable integer_var is not a table"),
		},
		{
			name: "error: table not found",
			joinSpec: &spec.TableJoin{
				LeftTable:   "unknown_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				On:          "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("table unknown_table is not declared"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := TableJoinOp{
				tableJoinSpec: tt.joinSpec,
			}

			err := c.Execute(context.Background(), tt.env)
			if tt.wantErr {
				assert.EqualError(t, err, tt.expError.Error())
				return
			}

			assert.NoError(t, err)
			for varName, varValue := range tt.expVariables {
				switch v := varValue.(type) {
				case time.Time:
					assert.True(t, v.Sub(tt.env.symbolRegistry[varName].(time.Time)) < time.Second)
				default:
					assert.Equal(t, v, tt.env.symbolRegistry[varName])
				}
			}
		})
	}
}
