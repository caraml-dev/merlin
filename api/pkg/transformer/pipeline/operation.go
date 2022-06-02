package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

type Op interface {
	Execute(context context.Context, environment *Environment) error
	AddInputOutput(input, output map[string]interface{}) error
	GetOperationTracingDetail() ([]types.TracingDetail, error)
}

type OperationIO struct {
	Name  string
	Value interface{}
}

type OperationTracing struct {
	mutex  sync.RWMutex
	Input  []map[string]interface{}
	Output []map[string]interface{}
	Specs  interface{}
	OpType types.OperationType
}

func NewOperationTracing(operationSpecs interface{}, opType types.OperationType) *OperationTracing {
	opTracing := &OperationTracing{
		Specs:  operationSpecs,
		OpType: opType,
	}
	return opTracing
}

func (ot *OperationTracing) AddInputOutput(input, output map[string]interface{}) error {
	ot.mutex.Lock()
	defer ot.mutex.Unlock()

	if ot.Input == nil {
		ot.Input = make([]map[string]interface{}, 0)
	}
	if ot.Output == nil {
		ot.Output = make([]map[string]interface{}, 0)
	}

	if err := sanitizeIO(input); err != nil {
		return err
	}
	if err := sanitizeIO(output); err != nil {
		return err
	}

	ot.Input = append(ot.Input, input)
	ot.Output = append(ot.Output, output)

	return nil
}

func sanitizeIO(io map[string]interface{}) error {
	for k, v := range io {
		tbl, ok := v.(*table.Table)
		if !ok {
			continue
		}
		formattedTable, err := table.TableToJson(tbl, spec.FromTable_RECORD)
		if err != nil {
			return err
		}
		io[k] = formattedTable
	}
	return nil
}

func (ot *OperationTracing) addInputOutputTable(input, output *table.Table) {
}

func (ot *OperationTracing) GetOperationTracingDetail() ([]types.TracingDetail, error) {
	refVal := reflect.ValueOf(ot.Specs)
	if refVal.Kind() == reflect.Slice {
		numOfSpecs := refVal.Len()
		if len(ot.Input) != numOfSpecs {
			return nil, fmt.Errorf("number of inputs is not match with number of specs")
		}
		if len(ot.Output) != numOfSpecs {
			return nil, fmt.Errorf("number of outputs is not match with number of specs")
		}
		result := make([]types.TracingDetail, numOfSpecs)
		for i := 0; i < numOfSpecs; i++ {
			result[i] = types.TracingDetail{
				Spec:   refVal.Index(i).Interface(),
				Input:  ot.Input[i],
				Output: ot.Output[i],
				OpType: ot.OpType,
			}
		}
		return result, nil
	}
	if len(ot.Input) != 1 {
		return nil, fmt.Errorf("input should has one record")
	}
	if len(ot.Output) != 1 {
		return nil, fmt.Errorf("output should has one record")
	}
	return []types.TracingDetail{
		{
			Spec:   ot.Specs,
			Input:  ot.Input[0],
			Output: ot.Output[0],
			OpType: ot.OpType,
		},
	}, nil
}
