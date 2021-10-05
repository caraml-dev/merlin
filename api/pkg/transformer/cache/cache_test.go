package cache

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/coocood/freecache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	testCases := []struct {
		desc string
		data interface{}
		key  []byte
	}{
		{
			desc: "Success - 1",
			data: map[string]string{"key": "value"},
			key:  []byte("key1"),
		},
		{
			desc: "Success - 2",
			data: map[string]string{"key": "value"},
			key:  []byte("key2"),
		},
		{
			desc: "Success - 3",
			data: []int{1, 2, 3},
			key:  []byte("key3"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dataByte, err := json.Marshal(tC.data)
			require.NoError(t, err)
			cache := NewInMemoryCache(1)
			cache.Insert(tC.key, dataByte, 1)
			cachedValue, err := cache.Fetch(tC.key)
			require.NoError(t, err)

			var val interface{}
			err = json.Unmarshal(cachedValue, &val)
			require.NoError(t, err)

			reflect.DeepEqual(tC.data, val)
		})
	}
}

func TestCache_Expiry(t *testing.T) {
	testCases := []struct {
		desc            string
		data            interface{}
		key             []byte
		delayOfFetching int
		err             error
	}{
		{
			desc:            "Success - Not expired",
			data:            map[string]string{"key": "value"},
			key:             []byte("key1"),
			delayOfFetching: 0,
		},
		{
			desc:            "Success - Expired",
			data:            map[string]string{"key": "value"},
			key:             []byte("key1"),
			delayOfFetching: 2,
			err:             freecache.ErrNotFound,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dataByte, err := json.Marshal(tC.data)
			require.NoError(t, err)
			cache := NewInMemoryCache(1)
			cache.Insert(tC.key, dataByte, 1*time.Second)

			if tC.delayOfFetching > 0 {
				time.Sleep(time.Duration(tC.delayOfFetching) * time.Second)
			}

			cachedValue, err := cache.Fetch(tC.key)
			assert.Equal(t, tC.err, err)

			if err == nil {
				var val interface{}
				err = json.Unmarshal(cachedValue, &val)
				require.NoError(t, err)

				reflect.DeepEqual(tC.data, val)
			}

		})
	}
}
