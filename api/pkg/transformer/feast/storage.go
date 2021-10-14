package feast

import (
	"context"
	"errors"
	"github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

func NewDirectStorageClient(storage *OnlineStorage, featureTables []spec.FeatureTable) (StorageClient, error) {
	switch storage.Storage.(type) {
	case *OnlineStorage_Redis:
		return RedisClient{
			host: storage.GetRedis().GetHost(),
			port: storage.GetRedis().GetPort(),
			featureTables: featureTables,
		}, nil
	}

	return nil, errors.New("unrecognized storage option")
}

type RedisClient struct {
	host string
	port int32
	featureTables []spec.FeatureTable
}

func (r RedisClient) GetOnlineFeatures(ctx context.Context, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	panic("implement me")
}
