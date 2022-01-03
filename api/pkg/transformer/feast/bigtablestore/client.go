package bigtablestore

import (
	"context"

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
	tables  map[string]*bigtable.Table
}

// NewClient instantiate new instance of BigTableClient
func NewClient(bigtableStorage *spec.BigTableStorage, featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) (*BigTableClient, error) {
	tables := make(map[string]*bigtable.Table)
	client, err := newBigtableClient(bigtableStorage)
	if err != nil {
		return nil, err
	}
	for _, featureSpec := range featureSpecs {
		btn := entityKeysToBigTable(featureSpec.Project, featureSpec.Entities)
		if _, exists := tables[btn]; !exists {
			tables[btn] = client.Open(btn)
		}
	}
	registry := newCachedCodecRegistry(tables)
	encoder := newEncoder(registry, featureSpecs, metadata)
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
	clientOpts = append(clientOpts, option.WithGRPCDialOption(grpc.WithKeepaliveParams(keepAliveParams)))
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
	rows := make([]bigtable.Row, 0)

	err = b.tables[query.table].ReadRows(ctx, query.rowList, func(row bigtable.Row) bool {
		rows = append(rows, row)
		return true
	}, bigtable.RowFilter(query.rowFilter))
	if err != nil {
		return nil, err
	}

	response, err := b.encoder.Decode(ctx, rows, req, query.entityKeys)
	if err != nil {
		return nil, err
	}

	return response, nil
}
