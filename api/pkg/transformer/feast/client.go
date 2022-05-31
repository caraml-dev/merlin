package feast

import (
	"fmt"
	"net"
	"strconv"

	feastSdk "github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/feast/bigtablestore"
	"github.com/gojek/merlin/pkg/transformer/feast/redis"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/pkg/errors"
)

func InitFeastServingClients(feastOptions Options, featureTableMetadata []*spec.FeatureTableMetadata, standardTransformerConfig *spec.StandardTransformerConfig) (Clients, error) {
	servingSources := getFeastServingSources(standardTransformerConfig)

	clients := Clients{}
	for _, source := range servingSources {
		feastSource := source
		if feastSource == spec.ServingSource_UNKNOWN {
			feastSource = feastOptions.DefaultFeastSource
		}
		feastClient, err := createFeastServingClient(feastOptions, featureTableMetadata, feastSource, standardTransformerConfig)
		if err != nil {
			return nil, err
		}
		clients[feastSource] = feastClient
	}
	return clients, nil
}

func createFeastServingClient(feastOptions Options, featureTableMetadata []*spec.FeatureTableMetadata, feastSource spec.ServingSource, standardTransformerConfig *spec.StandardTransformerConfig) (StorageClient, error) {
	storageConfig, ok := feastOptions.StorageConfigs[feastSource]
	if !ok {
		return nil, fmt.Errorf("not found storage for %s source", feastSource)
	}

	if storageConfig.ServingType == spec.ServingType_DIRECT_STORAGE {
		return newDirectStorageClient(storageConfig, featureTableMetadata, standardTransformerConfig)
	}

	var servingURL string
	switch storageConfig.Storage.(type) {
	case *spec.OnlineStorage_RedisCluster:
		servingURL = storageConfig.GetRedisCluster().FeastServingUrl
	case *spec.OnlineStorage_Redis:
		servingURL = storageConfig.GetRedis().FeastServingUrl
	case *spec.OnlineStorage_Bigtable:
		servingURL = storageConfig.GetBigtable().FeastServingUrl
	default:
		return nil, fmt.Errorf("not valid storage type")
	}
	return newFeastGrpcClient(servingURL)
}

func newFeastGrpcClient(url string) (*feastSdk.GrpcClient, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return nil, errors.Errorf("Unable to parse Feast Serving host (%s): %s", url, err)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.Errorf("Unable to parse Feast Serving port (%s): %s", url, err)
	}

	client, err := feastSdk.NewGrpcClient(host, portInt)
	if err != nil {
		return nil, errors.Errorf("Unable to initialize a Feast gRPC client: %s", err)
	}

	return client, nil
}

func newDirectStorageClient(storage *spec.OnlineStorage, featureTablesMetadata []*spec.FeatureTableMetadata, standardTransformerConfig *spec.StandardTransformerConfig) (StorageClient, error) {
	switch storage.Storage.(type) {
	case *spec.OnlineStorage_Redis:
		redisStorage := storage.GetRedis()
		return redis.NewRedisClient(redisStorage, featureTablesMetadata)
	case *spec.OnlineStorage_RedisCluster:
		redisClusterStorage := storage.GetRedisCluster()
		return redis.NewRedisClusterClient(redisClusterStorage, featureTablesMetadata)
	case *spec.OnlineStorage_Bigtable:
		bigtableStorage := storage.GetBigtable()
		featureTables := getFeatureTableSpecs(standardTransformerConfig)
		return bigtablestore.NewClient(bigtableStorage, featureTables, featureTablesMetadata)
	}
	return nil, errors.New("unrecognized storage option")
}
