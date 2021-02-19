package feast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/oliveagle/jsonpath"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/gojek/merlin/pkg/transformer"
)

var (
	feastError = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_serving_error_count",
		Help:      "The total number of error returned by feast serving",
	})

	feastLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_serving_request_duration_ms",
		Help:      "Feast serving latency histogram",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1,2,4,8,16,32,64,128,256,512,+Inf
	}, []string{"result"})

	feastFeatureStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: transformer.PromNamespace,
		Name:      "feast_feature_status_count",
		Help:      "Feature status by feature",
	}, []string{"feature", "status"})

	feastFeatureSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  transformer.PromNamespace,
		Name:       "feast_feature_value",
		Help:       "Summary of feature value",
		AgeBuckets: 1,
	}, []string{"feature"})
)

// Options for the Feast transformer.
type Options struct {
	ServingURL              string `envconfig:"FEAST_SERVING_URL" required:"true"`
	StatusMonitoringEnabled bool   `envconfig:"FEAST_FEATURE_STATUS_MONITORING_ENABLED" default:"false"`
	ValueMonitoringEnabled  bool   `envconfig:"FEAST_FEATURE_VALUE_MONITORING_ENABLED" default:"false"`
}

// Transformer wraps feast serving client to retrieve features.
type Transformer struct {
	feastClient      feast.Client
	config           *transformer.StandardTransformerConfig
	logger           *zap.Logger
	options          *Options
	defaultValues    map[string]*types.Value
	compiledJsonPath map[string]*jsonpath.Compiled
}

// NewTransformer initializes a new Transformer.
func NewTransformer(feastClient feast.Client, config *transformer.StandardTransformerConfig, options *Options, logger *zap.Logger) (*Transformer, error) {
	defaultValues := make(map[string]*types.Value)
	// populate default values
	for _, ft := range config.TransformerConfig.Feast {
		for _, f := range ft.Features {
			if len(f.DefaultValue) != 0 {
				feastValType := types.ValueType_Enum(types.ValueType_Enum_value[f.ValueType])
				defVal, err := getValue(f.DefaultValue, feastValType)
				if err != nil {
					logger.Warn(fmt.Sprintf("invalid default value for %s : %v, %v", f.Name, f.DefaultValue, err))
					continue
				}
				defaultValues[f.Name] = defVal
			}
		}
	}

	compiledJsonPath := make(map[string]*jsonpath.Compiled)
	for _, ft := range config.TransformerConfig.Feast {
		for _, configEntity := range ft.Entities {
			c, err := jsonpath.Compile(configEntity.JsonPath)
			if err != nil {
				return nil, fmt.Errorf("unable to compile jsonpath for entity %s: %s", configEntity.Name, configEntity.JsonPath)
			}
			compiledJsonPath[configEntity.JsonPath] = c
		}
	}

	return &Transformer{
		feastClient:      feastClient,
		config:           config,
		options:          options,
		logger:           logger,
		defaultValues:    defaultValues,
		compiledJsonPath: compiledJsonPath,
	}, nil
}

type FeastFeature struct {
	Columns []string        `json:"columns"`
	Data    [][]interface{} `json:"data"`
}

type result struct {
	tableName    string
	feastFeature *FeastFeature
	err          error
}

// Transform retrieves the Feast features values and add them into the request.
func (t *Transformer) Transform(ctx context.Context, request []byte) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.Transform")
	defer span.Finish()

	feastFeatures := make(map[string]*FeastFeature, len(t.config.TransformerConfig.Feast))

	// parallelize feast call per feature table
	resChan := make(chan result, len(t.config.TransformerConfig.Feast))
	for _, config := range t.config.TransformerConfig.Feast {
		go func(cfg *transformer.FeatureTable) {
			tableName := createTableName(cfg.Entities)
			val, err := t.getFeastFeature(ctx, tableName, request, cfg)
			resChan <- result{tableName, val, err}
		}(config)
	}

	// collect result
	for i := 0; i < cap(resChan); i++ {
		res := <-resChan
		if res.err != nil {
			return nil, res.err
		}
		feastFeatures[res.tableName] = res.feastFeature
	}

	out, err := enrichRequest(ctx, request, feastFeatures)
	if err != nil {
		return nil, err
	}

	return out, err
}

func (t *Transformer) getFeastFeature(ctx context.Context, tableName string, request []byte, config *transformer.FeatureTable) (*FeastFeature, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.getFeastFeature")
	span.SetTag("table.name", tableName)
	defer span.Finish()

	entities, err := t.buildEntitiesRequest(ctx, request, config.Entities)
	if err != nil {
		return nil, err
	}

	var features []string
	for _, feature := range config.Features {
		features = append(features, feature.Name)
	}

	feastRequest := feast.OnlineFeaturesRequest{
		Project:  config.Project,
		Entities: entities,
		Features: features,
	}
	t.logger.Debug("feast_request", zap.Any("feast_request", feastRequest))

	startTime := time.Now()
	feastResponse, err := t.feastClient.GetOnlineFeatures(ctx, &feastRequest)
	durationMs := time.Now().Sub(startTime).Milliseconds()
	if err != nil {
		feastLatency.WithLabelValues("error").Observe(float64(durationMs))
		feastError.Inc()
		return nil, err
	}
	feastLatency.WithLabelValues("success").Observe(float64(durationMs))

	t.logger.Debug("feast_response", zap.Any("feast_response", feastResponse.Rows()))

	feastFeature, err := t.buildFeastFeatures(ctx, feastResponse, config)
	if err != nil {
		return nil, err
	}

	return feastFeature, nil
}

func (t *Transformer) buildEntitiesRequest(ctx context.Context, request []byte, configEntities []*transformer.Entity) ([]feast.Row, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildEntitiesRequest")
	defer span.Finish()

	var entities []feast.Row
	var nodesBody interface{}
	err := json.Unmarshal(request, &nodesBody)
	if err != nil {
		return nil, err
	}

	for _, configEntity := range configEntities {
		c, ok := t.compiledJsonPath[configEntity.JsonPath]
		if !ok {
			c, err := jsonpath.Compile(configEntity.JsonPath)
			if err != nil {
				return nil, fmt.Errorf("unable to compile jsonpath for entity %s: %s", configEntity.Name, configEntity.JsonPath)
			}
			t.compiledJsonPath[configEntity.JsonPath] = c
		}

		vals, err := getValuesFromJSONPayload(nodesBody, configEntity, c)
		if err != nil {
			return nil, fmt.Errorf("unable to extract entity %s: %v", configEntity.Name, err)
		}

		if len(entities) == 0 {
			for _, val := range vals {
				entities = append(entities, feast.Row{
					configEntity.Name: val,
				})
			}
		} else {
			newEntities := []feast.Row{}
			for _, entity := range entities {
				for _, val := range vals {
					newFeastRow := feast.Row{}
					for k, v := range entity {
						newFeastRow[k] = v
					}

					newFeastRow[configEntity.Name] = val
					newEntities = append(newEntities, newFeastRow)
				}
			}
			entities = newEntities
		}
	}

	return entities, nil
}

func (t *Transformer) buildFeastFeatures(ctx context.Context, feastResponse *feast.OnlineFeaturesResponse, config *transformer.FeatureTable) (*FeastFeature, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.buildFeastFeatures")
	defer span.Finish()

	var columns []string
	for _, entity := range config.Entities {
		columns = append(columns, entity.Name)
	}
	for _, feature := range config.Features {
		columns = append(columns, feature.Name)
	}

	var data [][]interface{}
	status := feastResponse.Statuses()
	for i, feastRow := range feastResponse.Rows() {
		var row []interface{}
		for _, column := range columns {
			featureStatus := status[i][column]
			switch featureStatus {
			case serving.GetOnlineFeaturesResponse_PRESENT:
				rawValue := feastRow[column]
				featVal, err := getFeatureValue(rawValue)
				if err != nil {
					return nil, err
				}
				row = append(row, featVal)

				// put behind feature toggle since it will generate high cardinality metrics
				if t.options.ValueMonitoringEnabled {
					v, err := getFloatValue(featVal)
					if err != nil {
						continue
					}
					feastFeatureSummary.WithLabelValues(column).Observe(v)
				}
			case serving.GetOnlineFeaturesResponse_NOT_FOUND, serving.GetOnlineFeaturesResponse_NULL_VALUE, serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE:
				defVal, ok := t.defaultValues[column]
				if !ok {
					row = append(row, nil)
					continue
				}
				featVal, err := getFeatureValue(defVal)
				if err != nil {
					return nil, err
				}
				row = append(row, featVal)
			default:
				return nil, fmt.Errorf("Unsupported feature retrieval status: %s", featureStatus)
			}
			// put behind feature toggle since it will generate high cardinality metrics
			if t.options.StatusMonitoringEnabled {
				feastFeatureStatus.WithLabelValues(column, featureStatus.String()).Inc()
			}
		}
		data = append(data, row)
	}

	return &FeastFeature{
		Columns: columns,
		Data:    data,
	}, nil
}

func getFloatValue(val interface{}) (float64, error) {
	switch i := val.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	default:
		return math.NaN(), errors.New("getFloat: unknown value is of incompatible type")
	}
}

func createTableName(entities []*transformer.Entity) string {
	entityNames := make([]string, 0)
	for _, n := range entities {
		entityNames = append(entityNames, n.Name)
	}

	return strings.Join(entityNames, "_")
}

func getFeatureValue(val *types.Value) (interface{}, error) {
	switch val.Val.(type) {
	case *types.Value_StringVal:
		return val.GetStringVal(), nil
	case *types.Value_DoubleVal:
		return val.GetDoubleVal(), nil
	case *types.Value_FloatVal:
		return val.GetFloatVal(), nil
	case *types.Value_Int32Val:
		return val.GetInt32Val(), nil
	case *types.Value_Int64Val:
		return val.GetInt64Val(), nil
	case *types.Value_BoolVal:
		return val.GetBoolVal(), nil
	case *types.Value_StringListVal:
		return val.GetStringListVal(), nil
	case *types.Value_DoubleListVal:
		return val.GetDoubleListVal(), nil
	case *types.Value_FloatListVal:
		return val.GetFloatListVal(), nil
	case *types.Value_Int32ListVal:
		return val.GetInt32ListVal(), nil
	case *types.Value_Int64ListVal:
		return val.GetInt64ListVal(), nil
	case *types.Value_BoolListVal:
		return val.GetBoolListVal(), nil
	case *types.Value_BytesVal:
		return val.GetBytesVal(), nil
	case *types.Value_BytesListVal:
		return val.GetBytesListVal(), nil
	default:
		return nil, fmt.Errorf("unknown feature value type: %T", val.Val)
	}
}

func enrichRequest(ctx context.Context, request []byte, feastFeatures map[string]*FeastFeature) ([]byte, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "feast.enrichRequest")
	defer span.Finish()

	feastFeatureJSON, err := json.Marshal(feastFeatures)
	if err != nil {
		return nil, err
	}

	out, err := jsonparser.Set(request, feastFeatureJSON, transformer.FeastFeatureJSONField)
	if err != nil {
		return nil, err
	}

	return out, err
}
