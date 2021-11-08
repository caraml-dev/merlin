package feast

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/storage"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/go-redis/redis/v8"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/golang/protobuf/proto"
	"github.com/spaolacci/murmur3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewDirectStorageClient(storage *spec.OnlineStorage, featureTablesMetadata []*spec.FeatureTableMetadata) (StorageClient, error) {
	switch storage.Storage.(type) {
	case *spec.OnlineStorage_Redis:
		option := storage.GetRedis().GetOption()
		redisClient := redis.NewClient(&redis.Options{
			Addr:               storage.GetRedis().GetRedisAddress(),
			MaxRetries:         int(option.MaxRetries),
			MinRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
			MaxRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
			DialTimeout:        getNullableDuration(option.DialTimeout),
			ReadTimeout:        getNullableDuration(option.ReadTimeout),
			WriteTimeout:       getNullableDuration(option.WriteTimeout),
			PoolSize:           int(option.PoolSize),
			MaxConnAge:         getNullableDuration(option.MaxConnAge),
			PoolTimeout:        getNullableDuration(option.PoolTimeout),
			IdleTimeout:        getNullableDuration(option.IdleTimeout),
			IdleCheckFrequency: getNullableDuration(option.IdleCheckFrequency),
		})
		return RedisClient{
			encoder:   NewRedisEncoder(featureTablesMetadata),
			pipeliner: redisClient.Pipeline(),
		}, nil
	case *spec.OnlineStorage_RedisCluster:
		option := storage.GetRedis().GetOption()
		if option == nil {
			option = &spec.RedisOption{}
		}
		redisClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:              storage.GetRedisCluster().GetRedisAddress(),
			MaxRetries:         int(option.MaxRetries),
			MinRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
			MaxRetryBackoff:    getNullableDuration(option.MinRetryBackoff),
			DialTimeout:        getNullableDuration(option.DialTimeout),
			ReadTimeout:        getNullableDuration(option.ReadTimeout),
			WriteTimeout:       getNullableDuration(option.WriteTimeout),
			PoolSize:           int(option.PoolSize),
			MaxConnAge:         getNullableDuration(option.MaxConnAge),
			PoolTimeout:        getNullableDuration(option.PoolTimeout),
			IdleTimeout:        getNullableDuration(option.IdleTimeout),
			IdleCheckFrequency: getNullableDuration(option.IdleCheckFrequency),
		})
		return RedisClient{
			encoder:   NewRedisEncoder(featureTablesMetadata),
			pipeliner: redisClient.Pipeline(),
		}, nil
	}

	return nil, errors.New("unrecognized storage option")
}

func getNullableDuration(duration *durationpb.Duration) time.Duration {
	if duration == nil {
		return 0
	}
	return duration.AsDuration()
}

type RedisClient struct {
	encoder   RedisEncoder
	pipeliner redis.Pipeliner
}

func getMetadataKey(metadata *spec.FeatureTableMetadata) string {
	return metadataCompositeKey(metadata.Name, metadata.Project)
}

func metadataCompositeKey(featureTableName, project string) string {
	return fmt.Sprintf("%s-%s", project, featureTableName)
}

func NewRedisEncoder(featureTablesMetadata []*spec.FeatureTableMetadata) RedisEncoder {
	specs := make(map[string]*spec.FeatureTableMetadata)
	for _, metadata := range featureTablesMetadata {
		key := getMetadataKey(metadata)
		specs[key] = metadata
	}
	return RedisEncoder{specs: specs}
}

type RedisEncoder struct {
	specs map[string]*spec.FeatureTableMetadata
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

	key := &storage.RedisKeyV2{
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
	for index, encodedFeature := range encodedFeatures {
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

func (e RedisEncoder) buildFieldValues(entity feast.Row, project string, featureValues map[string]*types.Value, eventTimestamps map[string]*timestamppb.Timestamp) *serving.GetOnlineFeaturesResponse_FieldValues {
	entityFeatureValue := make(map[string]*types.Value)
	status := make(map[string]serving.GetOnlineFeaturesResponse_FieldStatus)
	for entityName, entityValue := range entity {
		entityFeatureValue[entityName] = entityValue
		status[entityName] = serving.GetOnlineFeaturesResponse_PRESENT
	}
	for featureReference, featureValue := range featureValues {
		entityFeatureValue[featureReference] = featureValue
		featureTable := getFeatureTableFromFeatureRef(featureReference)
		eventTimestamp := eventTimestamps[featureTable]
		maxAge := e.getFeatureTableMaxAge(featureTable, project)

		if proto.Equal(featureValue, &types.Value{}) {
			status[featureReference] = serving.GetOnlineFeaturesResponse_NOT_FOUND
		} else if maxAge > 0 && eventTimestamp.AsTime().Add(time.Duration(maxAge)*time.Second).Before(time.Now()) {
			status[featureReference] = serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
			entityFeatureValue[featureReference] = &types.Value{}
		} else {
			status[featureReference] = serving.GetOnlineFeaturesResponse_PRESENT
		}
	}
	return &serving.GetOnlineFeaturesResponse_FieldValues{
		Fields:   entityFeatureValue,
		Statuses: status,
	}
}

func (e RedisEncoder) DecodeStoredRedisValue(redisHashMaps [][]interface{}, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	fieldValues := make([]*serving.GetOnlineFeaturesResponse_FieldValues, len(redisHashMaps))
	for index, encodedHashMap := range redisHashMaps {
		decodedHashMap, eventTimestamps, err := e.decodeHashMap(encodedHashMap, req.Features)
		if err != nil {
			return nil, err
		}
		fieldValues[index] = e.buildFieldValues(req.Entities[index], req.Project, decodedHashMap, eventTimestamps)
	}
	return &feast.OnlineFeaturesResponse{
		RawResponse: &serving.GetOnlineFeaturesResponse{
			FieldValues: fieldValues,
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

func (e RedisEncoder) decodeHashMap(encodedHashMap []interface{}, featureReferences []string) (map[string]*types.Value, map[string]*timestamppb.Timestamp, error) {
	featureValues := make(map[string]*types.Value)
	for index, encodedField := range encodedHashMap[:len(featureReferences)] {
		featureValue, err := e.decodeFeature(encodedField)
		if err != nil {
			return nil, nil, err
		}
		featureValues[featureReferences[index]] = featureValue
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

func (r RedisClient) GetOnlineFeatures(ctx context.Context, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	encodedFeatureRequest, err := r.encoder.EncodeFeatureRequest(req)
	encodedEntities := encodedFeatureRequest.EncodedEntities
	encodedFeatures := encodedFeatureRequest.EncodedFeatures
	if err != nil {
		return nil, err
	}
	hmGetResults := make([]*redis.SliceCmd, len(encodedEntities))
	pipeline := r.pipeliner.Pipeline()
	for index, encodedEntity := range encodedEntities {
		hmGetResults[index] = pipeline.HMGet(ctx, encodedEntity, encodedFeatures...)
	}
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	redisHashMaps := make([][]interface{}, len(hmGetResults))
	for index, result := range hmGetResults {
		redisHashMap, err := result.Result()
		if err != nil {
			return nil, err
		}
		redisHashMaps[index] = redisHashMap
	}
	response, err := r.encoder.DecodeStoredRedisValue(redisHashMaps, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
