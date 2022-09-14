package redis

import (
	"context"
	"fmt"
	"github.com/feast-dev/feast/sdk/go/protos/feast/serving"
	"github.com/feast-dev/feast/sdk/go/protos/feast/types"
	"testing"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/go-redis/redis/v8"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestGetOnlineFeatures(t *testing.T) {
	testCases := []struct {
		desc                  string
		pipeliner             func() *pipelinerMock
		featureTablesMetadata []*spec.FeatureTableMetadata
		request               *feast.OnlineFeaturesRequest
		want                  *feast.OnlineFeaturesResponse
		wantError             bool
		err                   error
	}{
		{
			desc: "success",
			pipeliner: func() *pipelinerMock {
				pipeliner := &pipelinerMock{}
				firstResult := &redis.SliceCmd{}
				firstResult.SetVal([]interface{}{"\x18I", "\b\xe2\f"})
				secondResult := &redis.SliceCmd{}
				secondResult.SetVal([]interface{}{nil, nil})
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x01", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(firstResult)
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x02", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(secondResult)

				pipeliner.On("Exec", mock.Anything).Return(nil, nil)
				return pipeliner
			},
			request: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(1),
					},
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Project: "default",
			},
			featureTablesMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			want: &feast.OnlineFeaturesResponse{
				RawResponse: &serving.GetOnlineFeaturesResponseV2{
					Metadata: &serving.GetOnlineFeaturesResponseMetadata{
						FieldNames: &serving.FieldList{
							Val: []string{"driver_id", "driver_trips:trips_today"},
						},
					},
					Results: []*serving.GetOnlineFeaturesResponseV2_FieldVector{
						{
							Values:   []*types.Value{feast.Int64Val(1), feast.Int32Val(73)},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_PRESENT},
						},
						{
							Values:   []*types.Value{feast.Int64Val(2), {}},
							Statuses: []serving.FieldStatus{serving.FieldStatus_PRESENT, serving.FieldStatus_NOT_FOUND},
						},
					},
				},
			},
		},
		{
			desc: "exec command returning error",
			pipeliner: func() *pipelinerMock {
				pipeliner := &pipelinerMock{}
				firstResult := &redis.SliceCmd{}
				firstResult.SetVal([]interface{}{"\x18I", "\b\xe2\f"})
				secondResult := &redis.SliceCmd{}
				secondResult.SetVal([]interface{}{nil, nil})
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x01", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(firstResult)
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x02", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(secondResult)

				pipeliner.On("Exec", mock.Anything).Return(nil, fmt.Errorf("redis is down"))
				return pipeliner
			},
			request: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(1),
					},
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Project: "default",
			},
			featureTablesMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			wantError: true,
			err:       fmt.Errorf("redis is down"),
		},
		{
			desc: "Failed fetch the result",
			pipeliner: func() *pipelinerMock {
				pipeliner := &pipelinerMock{}
				firstResult := &redis.SliceCmd{}
				firstResult.SetVal([]interface{}{"\x18I", "\b\xe2\f"})
				secondResult := &redis.SliceCmd{}
				secondResult.SetErr(fmt.Errorf("failed to fetch value"))
				secondResult.SetVal([]interface{}{nil, nil})
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x01", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(firstResult)
				pipeliner.On("HMGet", mock.Anything, "\n\adefault\x12\tdriver_id\x1a\x02 \x02", "\xbe\xf9\x00\xf5", "_ts:driver_trips").Return(secondResult)
				pipeliner.On("Exec", mock.Anything).Return(nil, nil)
				return pipeliner
			},
			request: &feast.OnlineFeaturesRequest{
				Features: []string{"driver_trips:trips_today"},
				Entities: []feast.Row{
					{
						"driver_id": feast.Int64Val(1),
					},
					{
						"driver_id": feast.Int64Val(2),
					},
				},
				Project: "default",
			},
			featureTablesMetadata: []*spec.FeatureTableMetadata{
				{
					Name:    "driver_trips",
					Project: "default",
				},
			},
			wantError: true,
			err:       fmt.Errorf("failed to fetch value"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			redisClient, err := newClient(tC.pipeliner(), tC.featureTablesMetadata)
			if err != nil {
				panic(err)
			}
			got, err := redisClient.GetOnlineFeatures(context.Background(), tC.request)
			if tC.wantError {
				assert.Equal(t, tC.err, err)
			} else {
				if !proto.Equal(got.RawResponse, tC.want.RawResponse) {
					t.Errorf("expected %s, actual %s", tC.want.RawResponse, got.RawResponse)
				}
			}
		})
	}
}

type pipelinerMock struct {
	redis.Pipeliner
	mock.Mock
}

// HMGet provides a mock function with given fields: ctx, key, fields
func (_m *pipelinerMock) HMGet(ctx context.Context, key string, fields ...string) *redis.SliceCmd {
	_va := make([]interface{}, len(fields))
	for _i := range fields {
		_va[_i] = fields[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, key)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *redis.SliceCmd
	if rf, ok := ret.Get(0).(func(context.Context, string, ...string) *redis.SliceCmd); ok {
		r0 = rf(ctx, key, fields...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*redis.SliceCmd)
		}
	}

	return r0
}

func (_m *pipelinerMock) Exec(ctx context.Context) ([]redis.Cmder, error) {
	ret := _m.Called(ctx)

	var r0 []redis.Cmder
	if rf, ok := ret.Get(0).(func(context.Context) []redis.Cmder); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]redis.Cmder)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
