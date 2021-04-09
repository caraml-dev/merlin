package feast

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocks2 "github.com/gojek/merlin/pkg/transformer/cache/mocks"
	"github.com/gojek/merlin/pkg/transformer/types"
)

func TestFetchFeaturesFromCache(t *testing.T) {
	type mockCache struct {
		entity           feast.Row
		value            types.ValueRow
		errFetchingCache error
	}
	testCases := []struct {
		desc              string
		cacheMocks        []mockCache
		entities          []feast.Row
		project           string
		featuresFromCache types.ValueRows
		entityNotInCache  []feast.Row
	}{
		{
			desc:    "Success - 1, all entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: types.ValueRow{"1001", 1.1},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: types.ValueRow{"2002", 2.2},
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
			entityNotInCache: nil,
		},
		{
			desc:    "Success - 2, one of  entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: types.ValueRow{"1001", 1.1},
				},
				{
					entity: feast.Row{
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
			featuresFromCache: types.ValueRows{
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
			desc:    "Success - 3, none of entity has value in cache",
			project: "default",
			cacheMocks: []mockCache{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value:            nil,
					errFetchingCache: fmt.Errorf("Value not found"),
				},
				{
					entity: feast.Row{
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
			mockCache := &mocks2.Cache{}
			for _, cc := range tt.cacheMocks {
				key := CacheKey{Entity: cc.entity, Project: tt.project}
				keyByte, err := json.Marshal(key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.value)
				require.NoError(t, err)
				mockCache.On("Fetch", keyByte).Return(value, cc.errFetchingCache)

			}
			cached, notInCacheEntity := fetchFeaturesFromCache(mockCache, tt.entities, tt.project)
			assert.ElementsMatch(t, tt.featuresFromCache, cached)
			assert.Equal(t, tt.entityNotInCache, notInCacheEntity)
		})
	}
}

func TestInsertMultipleFeaturesToCache(t *testing.T) {
	type mockCache struct {
		entity            feast.Row
		value             types.ValueRow
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
					value:             types.ValueRow{"1001", 1.1},
					errInsertingCache: nil,
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             types.ValueRow{"2002", 2.2},
					errInsertingCache: nil,
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: types.ValueRow{"1001", 1.1},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: types.ValueRow{"2002", 2.2},
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
					value:             types.ValueRow{"1001", 1.1},
					errInsertingCache: nil,
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             types.ValueRow{"2002", 2.2},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: types.ValueRow{"1001", 1.1},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: types.ValueRow{"2002", 2.2},
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
					value:             types.ValueRow{"1001", 1.1},
					errInsertingCache: fmt.Errorf("Memory is full"),
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value:             types.ValueRow{"2002", 2.2},
					errInsertingCache: fmt.Errorf("Value is to big"),
				},
			},
			willBeCachedData: []entityFeaturePair{
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("1001"),
					},
					value: types.ValueRow{"1001", 1.1},
				},
				{
					entity: feast.Row{
						"driver_id": feast.StrVal("2002"),
					},
					value: types.ValueRow{"2002", 2.2},
				},
			},
			expectedError: fmt.Errorf("error inserting to cached: (value: [1001 1.1], with message: Memory is full),(value: [2002 2.2], with message: Value is to big)"),
		},
	}
	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			mockCache := &mocks2.Cache{}
			for _, cc := range tt.cacheMocks {
				key := CacheKey{Entity: cc.entity, Project: tt.project}
				keyByte, err := json.Marshal(key)
				require.NoError(t, err)
				value, err := json.Marshal(cc.value)
				require.NoError(t, err)
				mockCache.On("Insert", keyByte, value, mock.Anything).Return(cc.errInsertingCache)

			}
			err := insertMultipleFeaturesToCache(mockCache, tt.willBeCachedData, tt.project, 60*time.Second)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
