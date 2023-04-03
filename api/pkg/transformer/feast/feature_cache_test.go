package feast

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	feastTypes "github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/merlin/pkg/transformer/cache/mocks"
	"github.com/caraml-dev/merlin/pkg/transformer/types"
)

func TestFetchFeaturesFromCache(t *testing.T) {
	type mockCache struct {
		cacheKey         feast.Row
		cacheValue       *CacheValue
		errFetchingCache error
	}
	testCases := []struct {
		desc              string
		cacheMocks        []mockCache
		entities          []feast.Row
		project           string
		featuresFromCache types.ValueRows
		columnTypes       []feastTypes.ValueType_Enum
		entityNotInCache  []feast.Row
	}{
		{
			desc:    "Success - 1, all entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"1001", 1.1},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
				},
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"2002", 2.2},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
				},
			},
			entities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
			featuresFromCache: types.ValueRows{
				{
					"1001",
					1.1,
				},
				{
					"2002",
					2.2,
				},
			},
			columnTypes:      []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
			entityNotInCache: nil,
		},
		{
			desc:    "Success - 2, one of  entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"1001", 1.1},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
				},
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue:       nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
			},
			entities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
			featuresFromCache: types.ValueRows{
				{
					"1001",
					1.1,
				},
			},
			columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
			entityNotInCache: []feast.Row{
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
		},
		{
			desc:    "Success - 3, none of entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue:       nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
				{
					cacheKey: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue:       nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
			},
			entities: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
			featuresFromCache: nil,
			entityNotInCache: []feast.Row{
				{
					"driver_id": feast.StrVal("1001"),
				},
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			mockCache := &mocks.Cache{}
			for _, cc := range tt.cacheMocks {
				key := CacheKey{Entity: cc.cacheKey, Project: tt.project}
				keyByte, err := json.Marshal(key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.cacheValue)
				require.NoError(t, err)
				mockCache.On("Fetch", keyByte).Return(value, cc.errFetchingCache)
			}
			cached, columnTypes, notInCacheEntity := fetchFeaturesFromCache(mockCache, tt.entities, tt.project)
			assert.ElementsMatch(t, tt.featuresFromCache, cached)
			assert.Equal(t, tt.entityNotInCache, notInCacheEntity)
			assert.Equal(t, tt.columnTypes, columnTypes)
		})
	}
}

func TestInsertMultipleFeaturesToCache(t *testing.T) {
	type mockCache struct {
		entity            feast.Row
		cacheValue        *CacheValue
		errInsertingCache error
	}

	testCases := []struct {
		desc             string
		willBeCachedData []entityFeaturePair
		project          string
		cacheMocks       []mockCache
		expectedError    error
	}{
		{
			desc:    "Success - all features are successfully inserted to cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"1001", 1.1},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: nil,
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"2002", 2.2},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: nil,
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:       types.ValueRow{"1001", 1.1},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:       types.ValueRow{"2002", 2.2},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
		},
		{
			desc:    "Success - one feature is failing inserted to cache",
			project: "sample",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"1001", 1.1},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: nil,
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"2002", 2.2},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:       types.ValueRow{"1001", 1.1},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:       types.ValueRow{"2002", 2.2},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			expectedError: fmt.Errorf("error inserting to cached: (value: [2002 2.2], with message: Value is to big)"),
		},
		{
			desc:    "Success - all features are failing inserted to cache",
			project: "test",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"1001", 1.1},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: fmt.Errorf("Memory is full"),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					cacheValue: &CacheValue{
						ValueRow:   types.ValueRow{"2002", 2.2},
						ValueTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
					},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:       types.ValueRow{"1001", 1.1},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:       types.ValueRow{"2002", 2.2},
					columnTypes: []feastTypes.ValueType_Enum{feastTypes.ValueType_STRING, feastTypes.ValueType_DOUBLE},
				},
			},
			expectedError: fmt.Errorf("error inserting to cached: (value: [1001 1.1], with message: Memory is full),(value: [2002 2.2], with message: Value is to big)"),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			mockCache := &mocks.Cache{}
			for _, cc := range tt.cacheMocks {
				key := CacheKey{Entity: cc.entity, Project: tt.project}
				keyByte, err := json.Marshal(key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.cacheValue)
				require.NoError(t, err)
				mockCache.On("Insert", keyByte, value, mock.Anything).Return(cc.errInsertingCache)

			}
			err := insertMultipleFeaturesToCache(mockCache, tt.willBeCachedData, tt.project, 60*time.Second)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
