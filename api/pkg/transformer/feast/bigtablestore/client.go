package bigtablestore

import (
	"context"

	"cloud.google.com/go/bigtable"

	feast "github.com/feast-dev/feast/sdk/go"
	"github.com/gojek/merlin/pkg/transformer/spec"
)

type BigTableClient struct {
	encoder *Encoder
	tables  map[string]*bigtable.Table
}

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
