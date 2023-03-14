package pipeline

import (
	"context"
	"fmt"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/jinzhu/copier"
)

// UPIPreprocessOutputOp operation to convert all the preprocess result into types.UPIPredictionRequest
type UPIPreprocessOutputOp struct {
	outputSpec *spec.UPIPreprocessOutput
	*OperationTracing
}

// NewUPIPreprocessOutputOp function to initialize new operation
func NewUPIPreprocessOutputOp(outputSpec *spec.UPIPreprocessOutput, tracingEnabled bool) *UPIPreprocessOutputOp {
	output := &UPIPreprocessOutputOp{outputSpec: outputSpec}
	if tracingEnabled {
		output.OperationTracing = NewOperationTracing(outputSpec, types.UPIPreprocessOutputOp)
	}
	return output
}

// Execute output operation
// The only fields that modified in this output are
// 1. transformer_input
// 2. prediction_table
func (up *UPIPreprocessOutputOp) Execute(ctx context.Context, env *Environment) error {
	request := env.symbolRegistry.RawRequest()
	enrichedRequest, valid := request.(*types.UPIPredictionRequest)
	if !valid {
		return fmt.Errorf("not valid type %T", request)
	}

	predictionTable, err := getUPITableFromName(up.outputSpec.PredictionTableName, env)
	if err != nil {
		return err
	}

	transformerInputTables := make([]*upiv1.Table, len(up.outputSpec.TransformerInputTableNames))
	for idx, tblName := range up.outputSpec.TransformerInputTableNames {
		tbl, err := getUPITableFromName(tblName, env)
		if err != nil {
			return err
		}
		transformerInputTables[idx] = tbl
	}

	transformerInput := &upiv1.TransformerInput{}
	transformerInput.Tables = transformerInputTables
	copiedRequest := &types.UPIPredictionRequest{}
	copier.CopyWithOption(copiedRequest, enrichedRequest, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	copiedRequest.PredictionTable = predictionTable
	copiedRequest.TransformerInput = transformerInput

	env.SetOutput(copiedRequest)

	if up.OperationTracing != nil {
		outputDetail, err := copiedRequest.ToMap()
		if err != nil {
			return err
		}
		if err := up.OperationTracing.AddInputOutput(nil, outputDetail); err != nil {
			return err
		}
	}

	return nil
}

func getUPITableFromName(name string, env *Environment) (*upiv1.Table, error) {
	if name == "" {
		return nil, nil
	}
	tbl, err := getTable(env, name)
	if err != nil {
		return nil, err
	}
	upiTbl, err := tbl.ToUPITable(name)
	if err != nil {
		return nil, err
	}
	return upiTbl, nil
}
