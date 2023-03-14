package pipeline

import (
	"context"
	"fmt"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/jinzhu/copier"
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

// Execute output operation
// The only fields that modified in this output is prediction_result_table
func (up *UPIPostprocessOutputOp) Execute(ctx context.Context, env *Environment) error {
	modelResponse := env.symbolRegistry.ModelResponse()
	upiModelResponse, valid := modelResponse.(*types.UPIPredictionResponse)
	if !valid {
		return fmt.Errorf("not valid type %T", modelResponse)
	}

	copiedResponse := &types.UPIPredictionResponse{}
	if err := copier.CopyWithOption(copiedResponse, upiModelResponse, copier.Option{IgnoreEmpty: true, DeepCopy: true}); err != nil {
		return err
	}
	predictionResultTable, err := getUPITableFromName(up.outputSpec.PredictionResultTableName, env)
	if err != nil {
		return err
	}
	copiedResponse.PredictionResultTable = predictionResultTable
	env.SetOutput(copiedResponse)
	if up.OperationTracing != nil {
		outputDetail, err := copiedResponse.ToMap()
		if err != nil {
			return err
		}
		return up.OperationTracing.AddInputOutput(nil, outputDetail)
	}
	return nil
}
