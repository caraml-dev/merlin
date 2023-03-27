package redis

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/caraml-dev/merlin/pkg/transformer/spec"
	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	feastStorage "github.com/feast-dev/feast/sdk/go/protos/feast/storage"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/spaolacci/murmur3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newRedisEncoder(featureTablesMetadata []*spec.FeatureTableMetadata) RedisEncoder {
	specs := make(map[string]*spec.FeatureTableMetadata)
	for _, metadata := range featureTablesMetadata {
		key := getMetadataKey(metadata)
		specs[key] = metadata
	}
	return RedisEncoder{specs: specs}
}

func getMetadataKey(metadata *spec.FeatureTableMetadata) string {
	return metadataCompositeKey(metadata.Name, metadata.Project)
}

func metadataCompositeKey(featureTableName, project string) string {
	return fmt.Sprintf("%s-%s", project, featureTableName)
}

type EncodedFeatureRequest struct {
	EncodedEntities []string
	EncodedFeatures []string
}

func (e RedisEncoder) EncodeFeatureRequest(req *feast.OnlineFeaturesRequest) (EncodedFeatureRequest, error) {
	encodedEntities := make([]string, len(req.Entities))
	for index, entity := range req.Entities {
		encodedEntity, err := e.encodeEntity(req.Project, entity)
		if err != nil {
			return EncodedFeatureRequest{}, err
		}
		encodedEntities[index] = encodedEntity
	}
	encodedFeatures := e.encodeFeatureReferences(req.Features)
	return EncodedFeatureRequest{
		EncodedEntities: encodedEntities,
		EncodedFeatures: encodedFeatures,
	}, nil
}

func (e RedisEncoder) encodeEntity(project string, entity feast.Row) (string, error) {
	entityNames := make([]string, 0)
	for entityName := range entity {
		entityNames = append(entityNames, entityName)
	}
	sort.Strings(entityNames)
	entityValues := make([]*types.Value, len(entityNames))
	for indexEntityName, entityName := range entityNames {
		entityValues[indexEntityName] = entity[entityName]
	}

	key := &feastStorage.RedisKeyV2{
		Project:      project,
		EntityNames:  entityNames,
		EntityValues: entityValues,
	}
	keyByte, err := proto.Marshal(key)
	if err != nil {
		return "", err
	}
	return string(keyByte), nil
}

func (e RedisEncoder) encodeFeatureReferences(featureReferences []string) []string {
	encodedFeatures := make([]string, len(featureReferences))
	for index, featureReference := range featureReferences {
		hashedFeatureReference := murmur3.Sum32([]byte(featureReference))
		arr := make([]byte, 4)
		binary.LittleEndian.PutUint32(arr, hashedFeatureReference)
		encodedFeatures[index] = string(arr)
	}
	encodedTimestamps := e.encodeTimestamp(featureReferences)
	serializedFeatureWithTimestamp := make([]string, len(encodedFeatures)+len(encodedTimestamps))
	for index, encodedFeature := range encodedFeatures { // nolint: gosimple
		serializedFeatureWithTimestamp[index] = encodedFeature
	}
	for index, encodedTimestamp := range encodedTimestamps {
		serializedFeatureWithTimestamp[len(encodedFeatures)+index] = encodedTimestamp
	}

	return serializedFeatureWithTimestamp
}

func (e RedisEncoder) encodeTimestamp(featureReferences []string) []string {
	sortedFeatureTableSet := e.getSortedFeatureTableSet(featureReferences)
	encodedTimestamps := make([]string, len(sortedFeatureTableSet))
	for index, featureTable := range sortedFeatureTableSet {
		encodedTimestamps[index] = fmt.Sprintf("_ts:%s", featureTable)
	}
	return encodedTimestamps
}

func (e RedisEncoder) getSortedFeatureTableSet(featureReferences []string) []string {
	featureTableSet := make(map[string]bool)
	sortedFeatureTables := make([]string, 0)
	for _, featureReference := range featureReferences {
		featureTable := getFeatureTableFromFeatureRef(featureReference)
		if _, exists := featureTableSet[featureTable]; !exists {
			sortedFeatureTables = append(sortedFeatureTables, featureTable)
		}
		featureTableSet[featureTable] = true
	}
	return sortedFeatureTables
}

func (e RedisEncoder) getFeatureTableMaxAge(featureTable, project string) int64 {
	lookupKey := metadataCompositeKey(featureTable, project)
	return e.specs[lookupKey].MaxAge.GetSeconds()
}

func (e RedisEncoder) buildFieldVector(sortedEntityFieldName []string, featureReference []string, entity feast.Row, project string, featureValues []*types.Value, eventTimestamps map[string]*timestamppb.Timestamp) *serving.GetOnlineFeaturesResponseV2_FieldVector {
	entityFeatureValue := make([]*types.Value, len(sortedEntityFieldName)+len(featureValues))
	status := make([]serving.FieldStatus, len(entityFeatureValue))
	cnt := 0
	for _, field := range sortedEntityFieldName {
		entityFeatureValue[cnt] = entity[field]
		status[cnt] = serving.FieldStatus_PRESENT
		cnt++
	}
	for index, featureValue := range featureValues {
		entityFeatureValue[cnt] = featureValue
		featureTable := getFeatureTableFromFeatureRef(featureReference[index])
		eventTimestamp := eventTimestamps[featureTable]
		maxAge := e.getFeatureTableMaxAge(featureTable, project)

		if proto.Equal(featureValue, &types.Value{}) {
			status[cnt] = serving.FieldStatus_NOT_FOUND
		} else if maxAge > 0 && eventTimestamp.AsTime().Add(time.Duration(maxAge)*time.Second).Before(time.Now()) {
			status[cnt] = serving.FieldStatus_OUTSIDE_MAX_AGE
			entityFeatureValue[cnt] = &types.Value{}
		} else {
			status[cnt] = serving.FieldStatus_PRESENT
		}
		cnt++
	}
	return &serving.GetOnlineFeaturesResponseV2_FieldVector{
		Values:   entityFeatureValue,
		Statuses: status,
	}
}

func (e RedisEncoder) DecodeStoredRedisValue(redisHashMaps [][]interface{}, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	sortedEntityFieldNames := make([]string, len(req.Entities[0]))
	cnt := 0
	for fieldName := range req.Entities[0] {
		sortedEntityFieldNames[cnt] = fieldName
		cnt++
	}
	sort.Strings(sortedEntityFieldNames)

	fieldVectors := make([]*serving.GetOnlineFeaturesResponseV2_FieldVector, len(redisHashMaps))
	for index, encodedHashMap := range redisHashMaps {
		decodedValues, eventTimestamps, err := e.decodeHashMap(encodedHashMap, req.Features)
		if err != nil {
			return nil, err
		}
		fieldVectors[index] = e.buildFieldVector(sortedEntityFieldNames, req.Features, req.Entities[index], req.Project, decodedValues, eventTimestamps)
	}
	return &feast.OnlineFeaturesResponse{
		RawResponse: &serving.GetOnlineFeaturesResponseV2{
			Metadata: &serving.GetOnlineFeaturesResponseMetadata{
				FieldNames: &serving.FieldList{
					Val: append(sortedEntityFieldNames, req.Features...),
				},
			},
			Results: fieldVectors,
		},
	}, nil
}

func (e RedisEncoder) decodeFeature(encodedFeature interface{}) (*types.Value, error) {
	value := types.Value{}
	if encodedFeature != nil {
		err := proto.Unmarshal([]byte(encodedFeature.(string)), &value)
		if err != nil {
			return nil, err
		}
	}
	return &value, nil
}

func (e RedisEncoder) decodeEventTimestamp(encodedTimestamp interface{}) (*timestamppb.Timestamp, error) {
	eventTimestamp := timestamppb.Timestamp{}
	if encodedTimestamp != nil {
		err := proto.Unmarshal([]byte(encodedTimestamp.(string)), &eventTimestamp)
		if err != nil {
			return nil, err
		}
	}
	return &eventTimestamp, nil
}

func (e RedisEncoder) decodeHashMap(encodedHashMap []interface{}, featureReferences []string) ([]*types.Value, map[string]*timestamppb.Timestamp, error) {
	featureValues := make([]*types.Value, len(featureReferences))
	for index, encodedField := range encodedHashMap[:len(featureReferences)] {
		featureValue, err := e.decodeFeature(encodedField)
		if err != nil {
			return nil, nil, err
		}
		featureValues[index] = featureValue
	}

	eventTimestamps := make(map[string]*timestamppb.Timestamp)
	sortedFeatureTableSet := e.getSortedFeatureTableSet(featureReferences)
	for index, encodedTimestamp := range encodedHashMap[len(featureReferences):] {
		timestamp, err := e.decodeEventTimestamp(encodedTimestamp)
		if err != nil {
			return nil, nil, err
		}

		eventTimestamps[sortedFeatureTableSet[index]] = timestamp
	}

	return featureValues, eventTimestamps, nil
}

func getFeatureTableFromFeatureRef(ref string) string {
	return strings.Split(ref, ":")[0]
}
