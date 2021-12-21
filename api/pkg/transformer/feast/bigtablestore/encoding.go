package bigtablestore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigtable"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types/converter"
	"github.com/golang/protobuf/proto"
	"github.com/linkedin/goavro/v2"
)

// Encoder is used to encode OnlineFeaturesRequest into RowQuery
type Encoder struct {
	registry        CodecRegistry
	featureSpecs    map[featureTableKey]*spec.FeatureTable
	featureMetadata map[featureTableKey]*spec.FeatureTableMetadata
}

type featureTableKey struct {
	project string
	table   string
}

func entityKeysToBigTable(project string, entityKeys []*spec.Entity) string {
	keyNames := make([]string, len(entityKeys))
	for i, e := range entityKeys {
		keyNames[i] = e.Name
	}
	return project + "__" + strings.Join(keyNames, "__")
}

func feastValueToStringRepr(val *types.Value) (string, error) {
	switch v := val.Val.(type) {
	case *types.Value_StringVal:
		return val.GetStringVal(), nil
	case *types.Value_Int64Val:
		return strconv.FormatInt(val.GetInt64Val(), 10), nil
	case *types.Value_Int32Val:
		return strconv.FormatInt(int64(val.GetInt32Val()), 10), nil
	case *types.Value_BytesVal:
		return string(val.GetBytesVal()), nil
	default:
		return "", fmt.Errorf("unsupported value type %v", v)
	}
}

func feastRowToBigTableKey(entity feast.Row, entityKeys []*spec.Entity) (string, error) {
	entityValues := make([]string, len(entityKeys))
	for i, key := range entityKeys {
		entityValue, err := feastValueToStringRepr(entity[key.Name])
		if err != nil {
			return "", err
		}
		entityValues[i] = entityValue
	}
	return strings.Join(entityValues, "#"), nil
}

func compareEntityKeys(entityKeys1 []*spec.Entity, entityKeys2 []*spec.Entity) bool {
	if len(entityKeys1) != len(entityKeys2) {
		return false
	}
	for i := 0; i < len(entityKeys1); i++ {
		if !proto.Equal(entityKeys1[i], entityKeys2[i]) {
			return false
		}
	}
	return true
}

func getFeatureType(featureSpec *spec.FeatureTable, featureName string) (string, error) {
	for _, f := range featureSpec.Features {
		if f.Name == featureName {
			return f.ValueType, nil
		}
	}
	return "", fmt.Errorf("feature %s is not part of specification", featureName)
}

func avroToValueConversion(avroValue interface{}, featureType string) (*types.Value, error) {
	if avroValue == nil {
		return &types.Value{}, nil
	}
	switch strings.ToUpper(featureType) {
	case types.ValueType_STRING.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["string"], types.ValueType_STRING)
	case types.ValueType_INT32.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["int"], types.ValueType_INT32)
	case types.ValueType_INT64.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["long"], types.ValueType_INT64)
	case types.ValueType_FLOAT.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["float"], types.ValueType_FLOAT)
	case types.ValueType_DOUBLE.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["double"], types.ValueType_DOUBLE)
	case types.ValueType_BOOL.String():
		return converter.ToFeastValue(avroValue.(map[string]interface{})["boolean"], types.ValueType_BOOL)
	case types.ValueType_STRING_LIST.String():
		avroRecords := avroValue.(map[string]interface{})["array"].([]interface{})
		recordValues := make([]string, len(avroRecords))
		for i, r := range avroRecords {
			stringVal, _ := converter.ToString(r.(map[string]interface{})["string"])
			recordValues[i] = stringVal
		}
		return &types.Value{Val: &types.Value_StringListVal{
			StringListVal: &types.StringList{
				Val: recordValues,
			},
		}}, nil
	case types.ValueType_INT32_LIST.String():
		avroRecords := avroValue.(map[string]interface{})["array"].([]interface{})
		recordValues := make([]int32, len(avroRecords))
		for i, r := range avroRecords {
			intVal, err := converter.ToInt32(r.(map[string]interface{})["int"])
			if err != nil {
				return nil, err
			}
			recordValues[i] = intVal
		}
		return &types.Value{Val: &types.Value_Int32ListVal{
			Int32ListVal: &types.Int32List{
				Val: recordValues,
			},
		}}, nil
	case types.ValueType_INT64_LIST.String():
		avroRecords := avroValue.(map[string]interface{})["array"].([]interface{})
		recordValues := make([]int64, len(avroRecords))
		for i, r := range avroRecords {
			int64Val, err := converter.ToInt64(r.(map[string]interface{})["long"])
			if err != nil {
				return nil, err
			}
			recordValues[i] = int64Val
		}
		return &types.Value{Val: &types.Value_Int64ListVal{
			Int64ListVal: &types.Int64List{
				Val: recordValues,
			},
		}}, nil
	case types.ValueType_FLOAT_LIST.String():
		avroRecords := avroValue.(map[string]interface{})["array"].([]interface{})
		recordValues := make([]float32, len(avroRecords))
		for i, r := range avroRecords {
			floatVal, err := converter.ToFloat32(r.(map[string]interface{})["float"])
			if err != nil {
				return nil, err
			}
			recordValues[i] = floatVal
		}
		return &types.Value{Val: &types.Value_FloatListVal{
			FloatListVal: &types.FloatList{
				Val: recordValues,
			},
		}}, nil
	case types.ValueType_DOUBLE_LIST.String():
		avroRecords := avroValue.(map[string]interface{})["array"].([]interface{})
		recordValues := make([]float64, len(avroRecords))
		for i, r := range avroRecords {
			doubleVal, err := converter.ToFloat64(r.(map[string]interface{})["double"])
			if err != nil {
				return nil, err
			}
			recordValues[i] = doubleVal
		}
		return &types.Value{Val: &types.Value_DoubleListVal{
			DoubleListVal: &types.DoubleList{
				Val: recordValues,
			},
		}}, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

// NewEncoder instantiates new instance of Encoder
func NewEncoder(registry CodecRegistry, tables []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *Encoder {
	tableByKey := make(map[featureTableKey]*spec.FeatureTable)
	for _, tbl := range tables {
		tableByKey[featureTableKey{
			project: tbl.Project,
			table:   tbl.TableName,
		}] = tbl
	}
	metadataByKey := make(map[featureTableKey]*spec.FeatureTableMetadata)
	for _, m := range metadata {
		metadataByKey[featureTableKey{
			project: m.Project,
			table:   m.Name,
		}] = m
	}
	return &Encoder{
		registry:        registry,
		featureSpecs:    tableByKey,
		featureMetadata: metadataByKey,
	}
}

// RowQuery contains all information necessary to retrieve bigtable rows
// corresponding to list of entity rows and feature table(s).
type RowQuery struct {
	table      string
	entityKeys []*spec.Entity
	rowList    *bigtable.RowList
	rowFilter  bigtable.Filter
}

func (e *Encoder) extractCommonEntityKeys(project string, featureTables []string) ([]*spec.Entity, error) {
	if len(featureTables) == 0 {
		return nil, errors.New("must have at least one feature table requested")
	}

	entityKeysPerTable := make([][]*spec.Entity, len(featureTables))
	for i, ft := range featureTables {
		featureSpec := e.featureSpecs[featureTableKey{
			project: project,
			table:   ft,
		}]
		entityKeysPerTable[i] = featureSpec.Entities
	}

	for i := 1; i < len(entityKeysPerTable); i++ {
		if !compareEntityKeys(entityKeysPerTable[i], entityKeysPerTable[i-1]) {
			return nil, errors.New("all feature requested must have the same entity keys")
		}
	}

	return entityKeysPerTable[0], nil
}

// Encode parse the request into RowQuery based on the following rules:
// - All feature tables having the same entity keys are stored in the
// same BigTable, in the form of <project>__<1st entity key>__<2nd entity key>...
// - The row keys of the BigTable are in the form of  <entity key value_1>#<entity key value 2>...
// - Feature values from the same feature table are stored in the same column in
// the same BigTable table.
func (e *Encoder) Encode(req *feast.OnlineFeaturesRequest) (RowQuery, error) {
	featureTables, err := UniqueFeatureTablesFromFeatureRef(req.Features)
	if err != nil {
		return RowQuery{}, err
	}
	entityKeys, err := e.extractCommonEntityKeys(req.Project, featureTables)
	if err != nil {
		return RowQuery{}, err
	}

	var encodedEntities bigtable.RowList
	for _, entity := range req.Entities {
		encodedEntity, err := feastRowToBigTableKey(entity, entityKeys)
		if err != nil {
			return RowQuery{}, err
		}
		encodedEntities = append(encodedEntities, encodedEntity)
	}

	return RowQuery{
		table:      entityKeysToBigTable(req.Project, entityKeys),
		entityKeys: entityKeys,
		rowList:    &encodedEntities,
		rowFilter:  bigtable.FamilyFilter(strings.Join(featureTables, "|")),
	}, nil
}

func (e *Encoder) decodeAvro(ctx context.Context, row bigtable.Row, project string, entityKeys []*spec.Entity) (map[featureTableKey]map[string]interface{}, map[featureTableKey]time.Time, error) {
	featureValues := make(map[featureTableKey]map[string]interface{})
	featureTimestamps := make(map[featureTableKey]time.Time)
	for featureTable, cells := range row {
		if len(cells) > 0 {
			sort.Slice(cells, func(i, j int) bool {
				return cells[i].Timestamp.Time().After(cells[j].Timestamp.Time())
			})
		}
		cell := cells[0]
		schemaRef := cell.Value[:4]
		codec, err := e.registry.GetCodec(ctx, schemaRef, project, entityKeys)
		if err != nil {
			return nil, nil, err
		}
		avroValues, _, err := codec.NativeFromBinary(cell.Value[4:])
		if err != nil {
			return nil, nil, err
		}
		ftKey := featureTableKey{
			project: project,
			table:   featureTable,
		}
		featureValues[ftKey] = avroValues.(map[string]interface{})
		featureTimestamps[ftKey] = cell.Timestamp.Time()
	}
	return featureValues, featureTimestamps, nil
}

// Decode decodes given BigTable rows into Feast feature values, based on
// the assumption that the column values are stored as the concatenation of
// schema reference hash (4 bytes) and feature values serialized in Avro format.
func (e *Encoder) Decode(ctx context.Context, rows []bigtable.Row, req *feast.OnlineFeaturesRequest, entityKeys []*spec.Entity) (*feast.OnlineFeaturesResponse, error) {
	avroValueByKey := make(map[string]map[featureTableKey]map[string]interface{})
	timestampByKey := make(map[string]map[featureTableKey]time.Time)
	for _, row := range rows {
		avroValues, timestamps, err := e.decodeAvro(ctx, row, req.Project, entityKeys)
		if err != nil {
			return nil, err
		}
		avroValueByKey[row.Key()] = avroValues
		timestampByKey[row.Key()] = timestamps
	}

	fieldValues := make([]*serving.GetOnlineFeaturesResponse_FieldValues, len(req.Entities))
	for i, entity := range req.Entities {
		bigtableKey, err := feastRowToBigTableKey(entity, entityKeys)
		if err != nil {
			return nil, err
		}

		fields := make(map[string]*types.Value)
		status := make(map[string]serving.GetOnlineFeaturesResponse_FieldStatus)
		for k, v := range entity {
			fields[k] = v
			status[k] = serving.GetOnlineFeaturesResponse_PRESENT
		}
		avroValues := avroValueByKey[bigtableKey]
		timestamp := timestampByKey[bigtableKey]

		if avroValues == nil {
			for _, fr := range req.Features {
				fields[fr] = &types.Value{}
				status[fr] = serving.GetOnlineFeaturesResponse_NOT_FOUND
			}
			fieldValues[i] = &serving.GetOnlineFeaturesResponse_FieldValues{
				Fields:   fields,
				Statuses: status,
			}
			continue
		}
		for _, fr := range req.Features {
			featureRef, err := ParseFeatureRef(fr)
			if err != nil {
				return nil, err
			}
			maxAge := e.featureMetadata[featureTableKey{
				project: req.Project,
				table:   featureRef.FeatureTable,
			}].MaxAge
			if maxAge != nil && maxAge.GetSeconds() > 0 && timestamp[featureTableKey{
				project: req.Project,
				table:   featureRef.FeatureTable,
			}].Add(time.Duration(maxAge.GetSeconds())*time.Second).Before(time.Now()) {
				fields[fr] = &types.Value{}
				status[fr] = serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
				fieldValues[i] = &serving.GetOnlineFeaturesResponse_FieldValues{
					Fields:   fields,
					Statuses: status,
				}
				continue
			}

			avroValue := avroValueByKey[bigtableKey][featureTableKey{
				project: req.Project,
				table:   featureRef.FeatureTable,
			}][featureRef.Feature]

			featureSpec := e.featureSpecs[featureTableKey{
				project: req.Project,
				table:   featureRef.FeatureTable,
			}]

			featureType, err := getFeatureType(featureSpec, featureRef.Feature)
			if err != nil {
				return nil, err
			}
			val, err := avroToValueConversion(avroValue, featureType)
			if err != nil {
				return nil, err
			}
			fields[fr] = val
			status[fr] = serving.GetOnlineFeaturesResponse_PRESENT
		}
		fieldValues[i] = &serving.GetOnlineFeaturesResponse_FieldValues{
			Fields:   fields,
			Statuses: status,
		}
	}

	return &feast.OnlineFeaturesResponse{
		RawResponse: &serving.GetOnlineFeaturesResponse{
			FieldValues: fieldValues,
		},
	}, nil
}

// CodecRegistry returns an Avro codec given a schema reference hash in byte array
type CodecRegistry interface {
	GetCodec(ctx context.Context, schemaRef []byte, project string, entityKeys []*spec.Entity) (*goavro.Codec, error)
}

// CachedCodecRegistry instantiates an Avro codec by retrieving the schema from
// BigTable. The Avro schema corresponds to a FeatureTable is always stored as
// a row within the same BigTable, where the key is the schema reference hash. The
// codec is cached to avoid repeated calls to BigTable.
type CachedCodecRegistry struct {
	codecs map[string]*goavro.Codec
	tables map[string]*bigtable.Table
	sync.RWMutex
}

// NewCachedCodecRegistry instantiates a new CachedCodecRegistry
func NewCachedCodecRegistry(tables map[string]*bigtable.Table) *CachedCodecRegistry {
	return &CachedCodecRegistry{
		codecs: make(map[string]*goavro.Codec),
		tables: tables,
	}
}

// GetCodec returns an Avro codec, either from the in-memory cache or built
// from the Avro schema specification.
func (r *CachedCodecRegistry) GetCodec(ctx context.Context, schemaRef []byte, project string, entityKeys []*spec.Entity) (*goavro.Codec, error) {
	r.RLock()
	if codec, exists := r.codecs[string(schemaRef)]; exists {
		r.RUnlock()
		return codec, nil
	}
	r.RUnlock()
	tableName := entityKeysToBigTable(project, entityKeys)
	schemaKey := fmt.Sprintf("schema#%s", string(schemaRef))
	schemaValue, err := r.tables[tableName].ReadRow(ctx, schemaKey)
	if err != nil {
		return nil, err
	}
	codec, err := goavro.NewCodec(string(schemaValue["metadata"][0].Value))
	if err != nil {
		return nil, err
	}
	r.Lock()
	r.codecs[string(schemaRef)] = codec
	r.Unlock()
	return codec, nil
}
