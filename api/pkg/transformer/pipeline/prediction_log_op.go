package pipeline

import (
	"context"
	"fmt"
	"sort"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
	"github.com/caraml-dev/merlin/pkg/transformer/types/table"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jinzhu/copier"
)

type PredictionLogOp struct {
	producer          PredictionLogProducer
	predictionLogSpec *spec.PredictionLogConfig
}

// NewPredictionLogOp create instance that handle producing prediction log
func NewPredictionLogOp(predictionLogSpec *spec.PredictionLogConfig, logProducer PredictionLogProducer) *PredictionLogOp {
	return &PredictionLogOp{
		producer:          logProducer,
		predictionLogSpec: predictionLogSpec,
	}
}

// ProducePredictionLog build prediction log and publish it to storage
func (pl *PredictionLogOp) ProducePredictionLog(ctx context.Context, predictionResult *types.PredictionResult, env *Environment) error {
	predictionLog, err := pl.buildPredictionLog(predictionResult, env)
	if err != nil {
		return err
	}
	return pl.producer.Produce(ctx, predictionLog)
}

var now = timestamppb.Now

func (pl *PredictionLogOp) buildPredictionLog(predictionResult *types.PredictionResult, env *Environment) (*upiv1.PredictionLog, error) {

	originalRequest := env.SymbolRegistry().RawRequest()
	request, validReq := originalRequest.(*types.UPIPredictionRequest)
	if !validReq {
		return nil, fmt.Errorf("type of prediction request is not valid: %T", originalRequest)
	}

	upiRequest := &types.UPIPredictionRequest{}
	if err := copier.CopyWithOption(upiRequest, request, copier.Option{IgnoreEmpty: true, DeepCopy: true}); err != nil {
		return nil, err
	}

	log := &upiv1.PredictionLog{}
	log.ModelName = predictionResult.Metadata.ModelName
	log.ModelVersion = predictionResult.Metadata.ModelVersion
	log.ProjectName = predictionResult.Metadata.Project
	log.TargetName = upiRequest.TargetName
	log.TableSchemaVersion = converter.TableSchemaV1
	if metadata := upiRequest.Metadata; metadata != nil {
		log.PredictionId = metadata.PredictionId
	}

	modelInput, err := pl.buildModelInput(upiRequest, env)
	if err != nil {
		return nil, err
	}
	log.Input = modelInput

	modelOutput, err := pl.buildModelOutput(predictionResult, env)
	if err != nil {
		return nil, err
	}
	log.Output = modelOutput
	log.RequestTimestamp = now()

	return log, nil
}

func (pl *PredictionLogOp) buildModelInput(request *types.UPIPredictionRequest, env *Environment) (*upiv1.ModelInput, error) {
	modelInput := &upiv1.ModelInput{}
	modelInput.PredictionContext = request.PredictionContext

	requestHeaders := env.SymbolRegistry().RawRequestHeaders()
	modelInput.Headers = buildLogHeaders(requestHeaders)
	if request.PredictionTable != nil {
		featuresTbl, err := converter.TableToStruct(request.PredictionTable, converter.TableSchemaV1)
		if err != nil {
			return nil, err
		}
		modelInput.FeaturesTable = featuresTbl
	}

	if pl.predictionLogSpec.EntitiesTable != "" {
		entitiesTable, err := tableToStruct(request, pl.predictionLogSpec.EntitiesTable, env, func(tbl *table.Table) error {
			if tbl.NRow() != len(request.PredictionTable.Rows) {
				return fmt.Errorf("entities table dimension is not the same with prediction table")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		modelInput.EntitiesTable = entitiesTable
	}
	if pl.predictionLogSpec.RawFeaturesTable != "" {
		rawFeaturesTable, err := tableToStruct(request, pl.predictionLogSpec.RawFeaturesTable, env, func(tbl *table.Table) error {
			if tbl.NRow() != len(request.PredictionTable.Rows) {
				return fmt.Errorf("raw features table dimension is not the same with prediction table")
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		modelInput.RawFeatures = rawFeaturesTable
	}
	return modelInput, nil
}

func buildLogHeaders(headers map[string]string) []*upiv1.Header {
	if len(headers) == 0 {
		return nil
	}
	logHeaders := make([]*upiv1.Header, 0, len(headers))
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		logHeaders = append(logHeaders, &upiv1.Header{
			Key:   k,
			Value: headers[k],
		})
	}
	return logHeaders
}

func (pl *PredictionLogOp) buildModelOutput(predictionResult *types.PredictionResult, env *Environment) (*upiv1.ModelOutput, error) {

	modelOutput := &upiv1.ModelOutput{}
	responseHeaders := env.SymbolRegistry().ModelResponseHeaders()
	modelOutput.Headers = buildLogHeaders(responseHeaders)

	if predictionResult.Error == nil {
		response, validType := predictionResult.Response.(*types.UPIPredictionResponse)
		if !validType {
			return nil, fmt.Errorf("type of prediction result is not valid: %T", predictionResult)
		}

		modelOutput.PredictionContext = response.PredictionContext
		resultTable, err := converter.TableToStruct(response.PredictionResultTable, converter.TableSchemaV1)
		if err != nil {
			return nil, err
		}
		modelOutput.PredictionResultsTable = resultTable
	} else {
		modelOutput.Message = predictionResult.Error.Error()
		if statusErr, valid := status.FromError(predictionResult.Error); valid {
			modelOutput.Status = uint32(statusErr.Code())
		}
	}

	return modelOutput, nil
}

func tableToStruct(request *types.UPIPredictionRequest, tableName string, env *Environment, validationFn func(tbl *table.Table) error) (*structpb.Struct, error) {
	tbl, err := getTable(env, tableName)
	if err != nil {
		return nil, err
	}
	if !tbl.ColumnExist(table.RowIDColumn) {
		return nil, fmt.Errorf("column %s must be exist in table %s", table.RowIDColumn, tableName)
	}
	if validationFn != nil {
		if err := validationFn(tbl); err != nil {
			return nil, err
		}
	}
	upiTable, err := tbl.ToUPITable(tableName)
	if err != nil {
		return nil, err
	}

	entitiesVal, err := converter.TableToStruct(upiTable, converter.TableSchemaV1)
	if err != nil {
		return nil, err
	}
	return entitiesVal, nil
}
