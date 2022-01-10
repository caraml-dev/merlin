package bigtablestore

import (
	"context"

	"cloud.google.com/go/bigtable"
)

type storage interface {
	readRows(ctx context.Context, rowList *bigtable.RowList, filter bigtable.Filter) ([]bigtable.Row, error)
	readRow(ctx context.Context, row string) (bigtable.Row, error)
}

type btStorage struct {
	table *bigtable.Table
}

func (bt *btStorage) readRows(ctx context.Context, rowList *bigtable.RowList, filter bigtable.Filter) ([]bigtable.Row, error) {
	rows := make([]bigtable.Row, 0)
	err := bt.table.ReadRows(ctx, rowList, func(row bigtable.Row) bool {
		rows = append(rows, row)
		return true
	}, bigtable.RowFilter(filter))
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (bt *btStorage) readRow(ctx context.Context, row string) (bigtable.Row, error) {
	return bt.table.ReadRow(ctx, row)
}
