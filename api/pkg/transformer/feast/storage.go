package feast

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/feast-dev/feast/sdk/go"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/storage"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/go-redis/redis/v8"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/golang/protobuf/proto"
	"github.com/spaolacci/murmur3"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"time"
)

func NewDirectStorageClient(storage *OnlineStorage, featureTable *spec.FeatureTable) (StorageClient, error) {
	switch storage.Storage.(type) {
	case *OnlineStorage_Redis:
		redisClient := redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%d", storage.GetRedis().GetHost(), storage.GetRedis().GetPort()),
		})
		return RedisClient{
			encoder: RedisEncoder{spec: featureTable},
			pipeliner: redisClient.Pipeline(),
		}, nil
	}

	return nil, errors.New("unrecognized storage option")
}

type RedisClient struct {
	encoder RedisEncoder
	pipeliner redis.Pipeliner
}

type RedisEncoder struct {
	spec *spec.FeatureTable
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
	encodedFeatures := e.encodeFeature(req.Features)
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

	key := &storage.RedisKeyV2 {
		Project: project,
		EntityNames: entityNames,
		EntityValues: entityValues,
	}
	keyByte, err := proto.Marshal(key)
	if err != nil {
		return "", err
	}
	return string(keyByte), nil
}

func (e RedisEncoder) encodeFeature(featureReferences []string) []string {
	serializedFeatureWithTimestamp := make([]string, len(featureReferences) + 1)
	for index, featureReference := range featureReferences {
		hashedFeatureReference := murmur3.Sum32([]byte(featureReference))
		arr := make([]byte, 4)
		binary.LittleEndian.PutUint32(arr, hashedFeatureReference)
		serializedFeatureWithTimestamp[index] = string(arr)
	}
	serializedFeatureWithTimestamp[len(featureReferences)] = fmt.Sprintf("_ts:%s", e.spec.TableName)

	return serializedFeatureWithTimestamp
}

func (e RedisEncoder) buildFieldValues(entity feast.Row, featureValues map[string]*types.Value, eventTimestamp *timestamppb.Timestamp) *serving.GetOnlineFeaturesResponse_FieldValues {
	entityFeatureValue := make(map[string]*types.Value)
	status := make(map[string]serving.GetOnlineFeaturesResponse_FieldStatus)
	for entityName, entityValue := range entity {
		entityFeatureValue[entityName] = entityValue
		status[entityName] = serving.GetOnlineFeaturesResponse_PRESENT
	}
	for featureReference, featureValue := range featureValues {
		entityFeatureValue[featureReference] = featureValue
		if e.spec.MaxAge > 0 && eventTimestamp.AsTime().Add(time.Duration(e.spec.MaxAge) * time.Second).Before(time.Now()) {
			status[featureReference] = serving.GetOnlineFeaturesResponse_OUTSIDE_MAX_AGE
		} else {
			status[featureReference] = serving.GetOnlineFeaturesResponse_PRESENT
		}
	}
	return &serving.GetOnlineFeaturesResponse_FieldValues{
		Fields: entityFeatureValue,
		Statuses: status,
	}
}

func (e RedisEncoder) DecodeStoredRedisValue(sliceCmds []*redis.SliceCmd, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	fieldValues := make([]*serving.GetOnlineFeaturesResponse_FieldValues, len(sliceCmds))
	for index, sliceCmd := range sliceCmds {
		encodedHashMap, err := sliceCmd.Result()
		if err != nil {
			return nil, err
		}
		decodedHashMap, eventTimestamp, err := e.decodeHashMap(encodedHashMap)
		if err != nil {
			return nil, err
		}
		fieldValues[index] = e.buildFieldValues(req.Entities[index], decodedHashMap, eventTimestamp)
	}
	return &feast.OnlineFeaturesResponse{
		RawResponse: &serving.GetOnlineFeaturesResponse{
			FieldValues: fieldValues,
		},
	}, nil
}

func (e RedisEncoder) decodeFeature(encodedFeature interface{}) (*types.Value, error){
	value := types.Value{}
	if encodedFeature != nil {
		err := proto.Unmarshal([]byte(encodedFeature.(string)), &value)
		if err != nil {
			return nil, err
		}
	}
	return &value, nil
}

func (e RedisEncoder) decodeEventTimestamp(encodedTimestamp interface{}) (*timestamppb.Timestamp, error){
	eventTimestamp := timestamppb.Timestamp{}
	if encodedTimestamp != nil {
		err := proto.Unmarshal([]byte(encodedTimestamp.(string)), &eventTimestamp)
		if err != nil {
			return nil, err
		}
	}
	return &eventTimestamp, nil
}

func (e RedisEncoder) decodeHashMap(encodedHashMap []interface{}) (map[string]*types.Value, *timestamppb.Timestamp, error) {
	featureValues := make(map[string]*types.Value)
	for index, encodedField := range encodedHashMap[:len(encodedHashMap) - 1] {
		featureValue, err := e.decodeFeature(encodedField)
		if err != nil {
			return nil, nil, err
		}
		featureValues[fmt.Sprintf("%s:%s", e.spec.TableName, e.spec.Features[index].Name)] = featureValue
	}
	eventTimestamp, err := e.decodeEventTimestamp(encodedHashMap[len(encodedHashMap) - 1])
	if err != nil {
		return nil, nil, err
	}
	return featureValues, eventTimestamp, nil
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
	response, err := r.encoder.DecodeStoredRedisValue(hmGetResults, req)
	if err != nil {
		return nil, err
	}
	return response, nil
}
