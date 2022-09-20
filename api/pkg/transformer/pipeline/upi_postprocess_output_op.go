package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
)

// UPIPostprocessOutputOp
type UPIPostprocessOutputOp struct {
	outputSpec *spec.UPIPostprocessOutput
	*OperationTracing
}

// NewUPIPostprocessOutputOp
func NewUPIPostprocessOutputOp(outputSpec *spec.UPIPostprocessOutput, tracingEnabled bool) *UPIPostprocessOutputOp {
	output := &UPIPostprocessOutputOp{
		outputSpec: outputSpec,
	}
	if tracingEnabled {
		output.OperationTracing = NewOperationTracing(outputSpec, types.UPIPostprocessOutputOp)
	}
	return output
}

func (up *UPIPostprocessOutputOp) Execute(ctx context.Context, env *Environment) error {
	modelResponse := env.symbolRegistry.ModelResponse()
	upiModelResponse, valid := modelResponse.(*types.UPIPredictionResponse)
	if !valid {
		return fmt.Errorf("not valid type %T", modelResponse)
	}

	copiedResponse := *upiModelResponse
	predictionResultTable, err := getUPITableFromName(up.outputSpec.PredictionResultTableName, env)
	if err != nil {
		return err
	}
	copiedResponse.PredictionResultTable = predictionResultTable
	env.SetOutput(&copiedResponse)
	if up.OperationTracing != nil {
		return up.OperationTracing.AddInputOutput(nil, map[string]any{
			"output": &copiedResponse,
		})
	}
	return nil
}
