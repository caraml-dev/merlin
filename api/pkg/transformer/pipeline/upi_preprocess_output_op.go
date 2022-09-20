package pipeline

import (
	"context"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

// UPIPreprocessOutputOp
type UPIPreprocessOutputOp struct {
	outputSpec *spec.UPIPreprocessOutput
	*OperationTracing
}

func NewUPIPreprocessOutputOp(outputSpec *spec.UPIPreprocessOutput, tracingEnabled bool) *UPIPreprocessOutputOp {
	output := &UPIPreprocessOutputOp{outputSpec: outputSpec}
	if tracingEnabled {
		output.OperationTracing = NewOperationTracing(outputSpec, types.UPIPreprocessOutputOp)
	}
	return output
}

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
	copiedRequest := *enrichedRequest
	copiedRequest.PredictionTable = predictionTable
	copiedRequest.TransformerInput = transformerInput

	env.SetOutput(&copiedRequest)

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
	upiTbl, err := table.ToUPITable(tbl, name)
	if err != nil {
		return nil, err
	}
	return upiTbl, nil
}
