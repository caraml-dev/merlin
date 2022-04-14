package bigtablestore

import (
	"context"
	"encoding/base64"

	"cloud.google.com/go/bigtable"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/spec"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// BigTableClient provides the means of retrieving Feast features directly from
// BigTable without going through Feast Serving.
type BigTableClient struct {
	encoder *Encoder
	tables  map[string]storage
}

// NewClient instantiate new instance of BigTableClient
func NewClient(bigtableStorage *spec.BigTableStorage, featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) (*BigTableClient, error) {
	tables := make(map[string]storage)
	client, err := newBigtableClient(bigtableStorage)
	if err != nil {
		return nil, err
	}
	for index, featureSpec := range featureSpecs {
		btn := entityKeysToBigTable(featureSpec.Project, metadata[index].Entities)
		if _, exists := tables[btn]; !exists {
			btTable := client.Open(btn)
			tableStorage := &btStorage{table: btTable}
			tables[btn] = tableStorage
		}
	}
	return newClient(tables, featureSpecs, metadata)
}

func newClient(tables map[string]storage, featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) (*BigTableClient, error) {
	registry := newCachedCodecRegistry(tables)
	encoder, err := newEncoder(registry, featureSpecs, metadata)
	if err != nil {
		return nil, err
	}

	return &BigTableClient{
		encoder: encoder,
		tables:  tables,
	}, nil
}

func newBigtableClient(storage *spec.BigTableStorage) (*bigtable.Client, error) {
	opt := storage.Option
	btConfig := bigtable.ClientConfig{AppProfile: storage.AppProfile}
	var clientOpts []option.ClientOption
	clientOpts = append(clientOpts, option.WithGRPCConnectionPool(int(opt.GrpcConnectionPool)))

	keepAliveParams := keepalive.ClientParameters{}
	if opt.KeepAliveInterval != nil {
		keepAliveParams.Time = opt.KeepAliveInterval.AsDuration()
	}
	if opt.KeepAliveTimeout != nil {
		keepAliveParams.Timeout = opt.KeepAliveTimeout.AsDuration()
	}

	clientOpts = append(clientOpts,
		option.WithGRPCDialOption(grpc.WithKeepaliveParams(keepAliveParams)),
	)

	if opt.CredentialJson != "" {
		credsByte, err := base64.StdEncoding.DecodeString(opt.CredentialJson)
		if err != nil {
			return nil, err
		}
		clientOpts = append(clientOpts, option.WithCredentialsJSON(credsByte))
	}
	client, err := bigtable.NewClientWithConfig(context.Background(), storage.Project, storage.Instance, btConfig, clientOpts...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// GetOnlineFeatures will first convert the request to RowQuery, which contains all
// information needed to perform a query to BigTable to retrieve the Feast features,
// stored in Avro format. The retrieved BigTable rows are then decoded back into
// the Feast feature values again.
func (b BigTableClient) GetOnlineFeatures(ctx context.Context, req *feast.OnlineFeaturesRequest) (*feast.OnlineFeaturesResponse, error) {
	query, err := b.encoder.Encode(req)
	if err != nil {
		return nil, err
	}

	rows, err := b.tables[query.table].readRows(ctx, query.rowList, query.rowFilter)
	if err != nil {
		return nil, err
	}

	response, err := b.encoder.Decode(ctx, rows, req, query.entityKeys)
	if err != nil {
		return nil, err
	}

	return response, nil
}
