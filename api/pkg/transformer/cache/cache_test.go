package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	testCases := []struct {
		desc string
		data interface{}
		key  string
	}{
		{
			desc: "Success - 1",
			data: map[string]string{"key": "value"},
			key:  "key1",
		},
		{
			desc: "Success - 2",
			data: map[string]string{"key": "value"},
			key:  "key2",
		},
		{
			desc: "Success - 3",
			data: []int{1, 2, 3},
			key:  "key3",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cache := NewInMemoryCache()
			cache.Insert(tC.key, tC.data, 2*time.Second)

			cachedValue, ok := cache.Fetch(tC.key)
			require.True(t, ok)

			assert.Equal(t, tC.data, cachedValue)
		})
	}
}

func TestCache_Expiry(t *testing.T) {
	testCases := []struct {
		desc            string
		data            interface{}
		key             string
		delayOfFetching int
		foundInCache    bool
	}{
		{
			desc:            "Success - Not expired",
			data:            map[string]string{"key": "value"},
			key:             "key1",
			delayOfFetching: 0,
			foundInCache:    true,
		},
		{
			desc:            "Success - Expired",
			data:            map[string]string{"key": "value"},
			key:             "key1",
			delayOfFetching: 3,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cache := NewInMemoryCache()
			cache.Insert(tC.key, tC.data, 2*time.Second)

			if tC.delayOfFetching > 0 {
				time.Sleep(time.Duration(tC.delayOfFetching) * time.Second)
			}

			cachedValue, ok := cache.Fetch(tC.key)
			require.Equal(t, tC.foundInCache, ok)

			if ok {
				assert.Equal(t, tC.data, cachedValue)
			}

		})
	}
}
