package pipeline

import (
	"context"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	mErrors "github.com/gojek/merlin/pkg/errors"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/gojek/merlin/pkg/transformer/types/table"
)

// UPIAutoloadingOp operation to load all the inputs from incoming request to Standard Transformer or model response
type UPIAutoloadingOp struct {
	pipelineType types.Pipeline
	*OperationTracing
}

// NewUPIAutoloadingOp function to initialize new UPIAutoloadingOp
func NewUPIAutoloadingOp(pipelineType types.Pipeline, tracingEnabled bool) *UPIAutoloadingOp {
	op := &UPIAutoloadingOp{
		pipelineType: pipelineType,
	}
	if tracingEnabled {
		op.OperationTracing = NewOperationTracing(nil, types.UPIAutoloadingOp)
	}
	return op
}

// Execute autoloading operation
// Contains 2 types of operation
//  1. Preprocess. Load transformer input table and variables
//  2. Postprocess. Load model prediction response table
func (ua *UPIAutoloadingOp) Execute(ctx context.Context, env *Environment) error {
	if ua.pipelineType == types.Preprocess {
		return ua.autoLoadingPreprocessInput(ctx, env)
	}
	return ua.autoloadingPostprocessInput(ctx, env)
}

func validateRequest(payload *types.UPIPredictionRequest) error {
	if payload.PredictionTable != nil {
		if err := validateUPITable(payload.PredictionTable); err != nil {
			return err
		}
	}

	if payload.TransformerInput == nil {
		return nil
	}

	for _, tbl := range payload.TransformerInput.Tables {
		if tbl == nil {
			continue
		}
		if err := validateUPITable(tbl); err != nil {
			return err
		}
	}

	for _, variable := range payload.TransformerInput.Variables {
		if variable.Name == "" {
			return mErrors.NewInvalidInputError("variable name must be specified")
		}
	}
	return nil
}

func validateUPITable(tbl *upiv1.Table) error {
	// if user defined row_id in the columns then it will fail
	// since row_id will be automated created from row entry
	// if table doesn't have name we will throw error
	if tbl.Name == "" {
		return mErrors.NewInvalidInputError("table name must be specified")
	}
	for _, col := range tbl.Columns {
		if col.Name == table.RowIDColumn {
			return mErrors.NewInvalidInputError("row_id column is reserved, user is not allowed to define explicitly row_id column")
		}
	}
	return nil
}

func (ua *UPIAutoloadingOp) autoLoadingPreprocessInput(ctx context.Context, env *Environment) error {
	requestPayload := env.symbolRegistry.RawRequest()
	upiRequestPayload, valid := requestPayload.(*types.UPIPredictionRequest)
	if !valid {
		return mErrors.NewInvalidInputError("raw request is not valid")
	}

	if err := validateRequest(upiRequestPayload); err != nil {
		return err
	}

	variableSymbols, err := generateVariableFromRequest(upiRequestPayload, env)
	if err != nil {
		return err
	}
	tableSymbols, err := generateTableFromRequest(upiRequestPayload, env)
	if err != nil {
		return err
	}
	mergedSymbols := mergeMap(variableSymbols, tableSymbols)
	for name, val := range mergedSymbols {
		env.SetSymbol(name, val)
	}
	if ua.OperationTracing != nil {
		if err := ua.OperationTracing.AddInputOutput(nil, mergedSymbols); err != nil {
			return err
		}
	}
	return nil
}

func mergeMap(left, right map[string]any) map[string]any {
	for k, v := range right {
		left[k] = v
	}
	return left
}

func generateVariableFromRequest(requestPayload *types.UPIPredictionRequest, env *Environment) (map[string]any, error) {
	variables := make(map[string]any)

	if requestPayload.TransformerInput == nil {
		return variables, nil
	}

	for _, variable := range requestPayload.TransformerInput.Variables {
		switch variable.Type {
		case upiv1.Type_TYPE_INTEGER:
			variables[variable.Name] = variable.IntegerValue
		case upiv1.Type_TYPE_DOUBLE:
			variables[variable.Name] = variable.DoubleValue
		case upiv1.Type_TYPE_STRING:
			variables[variable.Name] = variable.StringValue
		default:
			return nil, mErrors.NewInvalidInputErrorf("unknown type %T", variable)
		}
	}
	return variables, nil
}

func generateTableFromRequest(requestPayload *types.UPIPredictionRequest, env *Environment) (map[string]any, error) {
	tables := make(map[string]any)
	if requestPayload.PredictionTable != nil {
		predictionTable, err := table.NewFromUPITable(requestPayload.PredictionTable)
		if err != nil {
			return nil, err
		}
		tables[requestPayload.PredictionTable.Name] = predictionTable
	}

	if requestPayload.TransformerInput == nil {
		return tables, nil
	}

	for _, upiTbl := range requestPayload.TransformerInput.Tables {
		table, err := table.NewFromUPITable(upiTbl)
		if err != nil {
			return nil, err
		}
		tables[upiTbl.Name] = table
	}
	return tables, nil
}

func (ua *UPIAutoloadingOp) autoloadingPostprocessInput(ctx context.Context, env *Environment) error {
	modelResponsePayload := env.symbolRegistry.ModelResponse()
	upiResponsePayload, valid := modelResponsePayload.(*types.UPIPredictionResponse)
	if !valid {
		return fmt.Errorf("model response is not valid type %T", modelResponsePayload)
	}

	if err := validateUPITable(upiResponsePayload.PredictionResultTable); err != nil {
		return err
	}

	tableSymbols, err := generateTableFromModelResponse(upiResponsePayload, env)
	if err != nil {
		return err
	}
	for name, val := range tableSymbols {
		env.SetSymbol(name, val)
	}
	if ua.OperationTracing != nil {
		if err := ua.OperationTracing.AddInputOutput(nil, tableSymbols); err != nil {
			return err
		}
	}
	return nil
}

func generateTableFromModelResponse(modelResponse *types.UPIPredictionResponse, env *Environment) (map[string]any, error) {
	tbl, err := table.NewFromUPITable(modelResponse.PredictionResultTable)
	if err != nil {
		return nil, err
	}
	tables := map[string]any{
		modelResponse.PredictionResultTable.Name: tbl,
	}
	return tables, nil
}
