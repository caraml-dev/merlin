package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/transformer/jsonpath"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/expression"
	"github.com/caraml-dev/merlin/pkg/transformer/types/series"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
)

func TestTableJoinOp_Execute(t *testing.T) {
	compiledExpression := expression.NewStorage()
	compiledJsonPath := jsonpath.NewStorage()
	logger, _ := zap.NewDevelopment()

	customerTable := table.New(
		series.New([]interface{}{1, 2, 3, 4}, series.Int, "customer_id"),
		series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Elise"}, series.String, "name"),
		series.New([]interface{}{24, 25, 23, 20}, series.Int, "age"),
	)

	orderTable := table.New(
		series.New([]interface{}{111, 222, 333, 444}, series.Int, "order_id"),
		series.New([]interface{}{3, 3, 2, 4}, series.Int, "customer_id"),
		series.New([]interface{}{500, 10, 12, 24}, series.Int, "amount"),
	)

	tableA := table.New(
		series.New([]interface{}{1, 2, 3}, series.Int, "col_a"),
	)

	tableB := table.New(
		series.New([]interface{}{11, 22, 33}, series.Int, "col_b"),
	)

	table1 := table.New(
		series.New([]interface{}{1, 2, 3}, series.Int, "col_int"),
		series.New([]interface{}{"a", "b", "c"}, series.String, "col_string"),
		series.New([]interface{}{true, false, true}, series.Bool, "col_bool"),
	)

	table2 := table.New(
		series.New([]interface{}{1, 2, 3, 4, 5}, series.Int, "col_int"),
		series.New([]interface{}{"a", "b", "c", "d", "e"}, series.String, "col_string"),
		series.New([]interface{}{0.1, 0.2, 0.3, 0.4, 0.5}, series.Int, "col_float"),
	)

	table3 := table.New(
		series.New([]interface{}{100000000.0, 222222222.0, nil}, series.Float, "col_1"),
		series.New([]interface{}{"a", "b", "c"}, series.String, "col_2"),
		series.New([]interface{}{true, false, true}, series.Bool, "col_3"),
	)

	table4 := table.New(
		series.New([]interface{}{100000000, 200000000, nil}, series.Int, "col_1"),
		series.New([]interface{}{"aa", "bb", "cc"}, series.String, "col1_1"),
		series.New([]interface{}{0.1, 0.2, 0.3}, series.Float, "col1_2"),
	)

	env := NewEnvironment(&CompiledPipeline{
		compiledJsonpath:   compiledJsonPath,
		compiledExpression: compiledExpression,
	}, logger)

	env.SetSymbol("customer_table", customerTable)
	env.SetSymbol("order_table", orderTable)
	env.SetSymbol("table_a", tableA)
	env.SetSymbol("table_b", tableB)
	env.SetSymbol("table_1", table1)
	env.SetSymbol("table_2", table2)
	env.SetSymbol("table_3", table3)
	env.SetSymbol("table_4", table4)
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
			name: "success: left join two table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount")),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: right join two table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "order_table",
				RightTable:  "customer_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_RIGHT,
				OnColumn:    "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4, 1}, series.Int, "customer_id"),
					series.New([]interface{}{333, 111, 222, 444, nil}, series.Int, "order_id"),
					series.New([]interface{}{12, 500, 10, 24, nil}, series.Int, "amount"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise", "Ramesh"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20, 24}, series.Int, "age"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: right join two table -- join column has different types",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_3",
				RightTable:  "table_4",
				OutputTable: "table_temp",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "col_1",
			},
			env: env,
			expVariables: map[string]interface{}{
				"table_temp": table.New(
					series.New([]interface{}{100000000.0, 222222222.0, nil}, series.Float, "col_1"),
					series.New([]interface{}{"a", "b", "c"}, series.String, "col_2"),
					series.New([]interface{}{true, false, true}, series.Bool, "col_3"),
					series.New([]interface{}{"aa", nil, nil}, series.String, "col1_1"),
					series.New([]interface{}{0.1, nil, nil}, series.Float, "col1_2"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: inner join two table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_INNER,
				OnColumn:    "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: outer join two table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_OUTER,
				OnColumn:    "customer_id",
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: cross join two table (no onColumn)",
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
			name: "success: concat two table (no onColumn)",
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
			name: "error: join column does not exist in left table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "order_id",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column order_id does not exist in left table"),
		},
		{
			name: "error: join column does not exist in right table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column age does not exist in right table"),
		},
		{
			name: "error: not a table -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "integer_var",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: variable 'integer_var' is not a table"),
		},
		{
			name: "error: table not found -- join on one column (using deprecated onColumn field)",
			joinSpec: &spec.TableJoin{
				LeftTable:   "unknown_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumn:    "age",
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: table 'unknown_table' is not declared"),
		},
		{
			name: "success: left join two table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"customer_id"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount")),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: right join two table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "order_table",
				RightTable:  "customer_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_RIGHT,
				OnColumns:   []string{"customer_id"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4, 1}, series.Int, "customer_id"),
					series.New([]interface{}{333, 111, 222, 444, nil}, series.Int, "order_id"),
					series.New([]interface{}{12, 500, 10, 24, nil}, series.Int, "amount"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise", "Ramesh"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20, 24}, series.Int, "age"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: inner join two table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_INNER,
				OnColumns:   []string{"customer_id"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: outer join two table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_OUTER,
				OnColumns:   []string{"customer_id"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 3, 4}, series.Int, "customer_id"),
					series.New([]interface{}{"Ramesh", "Hardik", "Komal", "Komal", "Elise"}, series.String, "name"),
					series.New([]interface{}{24, 25, 23, 23, 20}, series.Int, "age"),
					series.New([]interface{}{nil, 333, 111, 222, 444}, series.Int, "order_id"),
					series.New([]interface{}{nil, 12, 500, 10, 24}, series.Int, "amount"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "error: join column does not exist in left table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"order_id"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column order_id does not exist in left table"),
		},
		{
			name: "error: join column does not exist in right table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "customer_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"age"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column age does not exist in right table"),
		},
		{
			name: "error: not a table -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "integer_var",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"age"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: variable 'integer_var' is not a table"),
		},
		{
			name: "error: table not found -- join on one column",
			joinSpec: &spec.TableJoin{
				LeftTable:   "unknown_table",
				RightTable:  "order_table",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"age"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: table 'unknown_table' is not declared"),
		},
		{
			name: "success: left join two table-- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_int", "col_string"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3}, series.Int, "col_int"),
					series.New([]interface{}{"a", "b", "c"}, series.String, "col_string"),
					series.New([]interface{}{true, false, true}, series.Bool, "col_bool"),
					series.New([]interface{}{0.1, 0.2, 0.3}, series.Int, "col_float"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: right join two table-- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_RIGHT,
				OnColumns:   []string{"col_int", "col_string"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 4, 5}, series.Int, "col_int"),
					series.New([]interface{}{"a", "b", "c", "d", "e"}, series.String, "col_string"),
					series.New([]interface{}{true, false, true, nil, nil}, series.Bool, "col_bool"),
					series.New([]interface{}{0.1, 0.2, 0.3, 0.4, 0.5}, series.Int, "col_float"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: inner join two table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_INNER,
				OnColumns:   []string{"col_int", "col_string"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3}, series.Int, "col_int"),
					series.New([]interface{}{"a", "b", "c"}, series.String, "col_string"),
					series.New([]interface{}{true, false, true}, series.Bool, "col_bool"),
					series.New([]interface{}{0.1, 0.2, 0.3}, series.Int, "col_float"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "success: outer join two table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_OUTER,
				OnColumns:   []string{"col_int", "col_string"},
			},
			env: env,
			expVariables: map[string]interface{}{
				"result_table": table.New(
					series.New([]interface{}{1, 2, 3, 4, 5}, series.Int, "col_int"),
					series.New([]interface{}{"a", "b", "c", "d", "e"}, series.String, "col_string"),
					series.New([]interface{}{true, false, true, nil, nil}, series.Bool, "col_bool"),
					series.New([]interface{}{0.1, 0.2, 0.3, 0.4, 0.5}, series.Int, "col_float"),
				),
			},
			wantErr:  false,
			expError: nil,
		},
		{
			name: "error: join column does not exist in left table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_float"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_float does not exist in left table"),
		},
		{
			name: "error: join column does not exist in right table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_bool"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_bool does not exist in right table"),
		},
		{
			name: "error: a join column does not exist in left table or right table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_float", "col_bool"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_float does not exist in left table and column col_bool does not exist in right table"),
		},
		{
			name: "error: join column does not exist in right table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_bool"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_bool does not exist in right table"),
		},
		{
			name: "error: join column does not exist in both tables -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_float32"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_float32 does not exist in left table and right table"),
		},
		{
			name: "error: join columns does not exist in both tables -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_float32", "col_float64"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_float32, col_float64 does not exist in left table and right table"),
		},
		{
			name: "error: multiple join columns does not exist in left table or right table -- join on multiple columns",
			joinSpec: &spec.TableJoin{
				LeftTable:   "table_1",
				RightTable:  "table_2",
				OutputTable: "result_table",
				How:         spec.JoinMethod_LEFT,
				OnColumns:   []string{"col_float", "col_bool", "col_float32", "col_float64"},
			},
			env:      env,
			wantErr:  true,
			expError: errors.New("invalid input: invalid join column: column col_float, col_float32, col_float64 does not exist in left table and column col_bool, col_float32, col_float64 does not exist in right table"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := TableJoinOp{
				tableJoinSpec:    tt.joinSpec,
				OperationTracing: NewOperationTracing(tt.joinSpec, types.TableJoinOpType),
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
