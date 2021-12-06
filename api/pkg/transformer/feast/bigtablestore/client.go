package bigtablestore

import (
	"context"

	"cloud.google.com/go/bigtable"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

// BigTableClient provides the means of retrieving Feast features directly from
// BigTable without going through Feast Serving.
type BigTableClient struct {
	encoder *Encoder
	tables  map[string]*bigtable.Table
}

// NewClient instantiate new instance of BigTableClient
func NewClient(client *bigtable.Client, project string, featureSpecs []*spec.FeatureTable, metadata []*spec.FeatureTableMetadata) *BigTableClient {
	tables := make(map[string]*bigtable.Table)
	for _, featureSpec := range featureSpecs {
		btn := entityKeysToBigTable(project, featureSpec.Entities)
		if _, exists := tables[btn]; !exists {
			tables[btn] = client.Open(btn)
		}
	}
	registry := NewCachedCodecRegistry(tables)
	encoder := NewEncoder(registry, featureSpecs, metadata)
	return &BigTableClient{
		encoder: encoder,
		tables:  tables,
	}
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
