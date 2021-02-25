package feast

import (
	"encoding/json"
	"fmt"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/feast/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFetchFeaturesFromCache(t *testing.T) {
	type mockCache struct {
		key              feast.Row
		value            FeatureData
		errFetchingCache error
	}
	testCases := []struct {
		desc              string
		cacheMocks        []mockCache
		entities          []feast.Row
		featuresFromCache FeaturesData
		entityNotInCache  []feast.Row
	}{
		{
			desc: "Success - 1, all entity has value in cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: FeatureData{"1001", 1.1},
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: FeatureData{"2002", 2.2},
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
			featuresFromCache: FeaturesData{
				{
					"1001",
					1.1,
				},
				{
					"2002",
					2.2,
				},
			},
			entityNotInCache: nil,
		},
		{
			desc: "Success - 2, one of  entity has value in cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: FeatureData{"1001", 1.1},
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:            nil,
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
			featuresFromCache: FeaturesData{
				{
					"1001",
					1.1,
				},
			},
			entityNotInCache: []feast.Row{
				{
					"driver_id": feast.StrVal("2002"),
				},
			},
		},
		{
			desc: "Success - 3, none of entity has value in cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:            nil,
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
				keyByte, err := json.Marshal(cc.key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.value)
				require.NoError(t, err)
				mockCache.On("Fetch", keyByte).Return(value, cc.errFetchingCache)

			}
			cached, notInCacheEntity := fetchFeaturesFromCache(mockCache, tt.entities)
			assert.ElementsMatch(t, tt.featuresFromCache, cached)
			assert.Equal(t, tt.entityNotInCache, notInCacheEntity)
		})
	}
}

func TestInsertMultipleFeaturesToCache(t *testing.T) {
	type mockCache struct {
		key               feast.Row
		value             FeatureData
		errInsertingCache error
	}
	testCases := []struct {
		desc             string
		willBeCachedData []cacheableFeatureData
		cacheMocks       []mockCache
		expectedError    error
	}{
		{
			desc: "Success - all features are successfully inserted to cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:             FeatureData{"1001", 1.1},
					errInsertingCache: nil,
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             FeatureData{"2002", 2.2},
					errInsertingCache: nil,
				},
			},
			willBeCachedData: []cacheableFeatureData{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: FeatureData{"1001", 1.1},
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: FeatureData{"2002", 2.2},
				},
			},
		},
		{
			desc: "Success - one feature is failing inserted to cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:             FeatureData{"1001", 1.1},
					errInsertingCache: nil,
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             FeatureData{"2002", 2.2},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []cacheableFeatureData{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: FeatureData{"1001", 1.1},
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: FeatureData{"2002", 2.2},
				},
			},
			expectedError: fmt.Errorf("error inserting to cached: (value: [2002 2.2], with message: Value is to big)"),
		},
		{
			desc: "Success - all features are failing inserted to cache",
			cacheMocks: []mockCache{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:             FeatureData{"1001", 1.1},
					errInsertingCache: fmt.Errorf("Memory is full"),
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             FeatureData{"2002", 2.2},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []cacheableFeatureData{
				{
					key: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: FeatureData{"1001", 1.1},
				},
				{
					key: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: FeatureData{"2002", 2.2},
				},
			},
			expectedError: fmt.Errorf("error inserting to cached: (value: [1001 1.1], with message: Memory is full),(value: [2002 2.2], with message: Value is to big)"),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			mockCache := &mocks.Cache{}
			for _, cc := range tt.cacheMocks {
				keyByte, err := json.Marshal(cc.key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.value)
				require.NoError(t, err)
				mockCache.On("Insert", keyByte, value, mock.Anything).Return(cc.errInsertingCache)

			}
			err := insertMultipleFeaturesToCache(mockCache, tt.willBeCachedData, 60)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
