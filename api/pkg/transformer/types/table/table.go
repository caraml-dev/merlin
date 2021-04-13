package table

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	gota "github.com/go-gota/gota/series"

	"github.com/gojek/merlin/pkg/transformer/types/series"
)

type Table struct {
	dataFrame *dataframe.DataFrame
}

func NewTable(df *dataframe.DataFrame) *Table {
	return &Table{dataFrame: df}
}

func New(se ...*series.Series) *Table {
	ss := make([]gota.Series, 0, 0)
	for _, gs := range se {
		ss = append(ss, *gs.Series())
	}

	df := dataframe.New(ss...)
	return &Table{dataFrame: &df}
}

func (t *Table) Row(row int) (*Table, error) {
	if row < 0 || row >= t.dataFrame.Nrow() {
		return nil, fmt.Errorf("invalid row number, expected: 0 <= row < %d, got: %d", t.dataFrame.Nrow(), row)
	}

	subsetDataframe := t.dataFrame.Subset(row)
	if subsetDataframe.Err != nil {
		return nil, subsetDataframe.Err
	}

	return &Table{&subsetDataframe}, nil
}

func (t *Table) Col(colName string) (*series.Series, error) {
	s := t.dataFrame.Col(colName)
	if s.Err != nil {
		return nil, s.Err
	}

	return series.NewSeries(&s), nil
}

func (t *Table) DataFrame() *dataframe.DataFrame {
	return t.dataFrame
}

func (t *Table) Copy() *Table {
	df := t.dataFrame.Copy()
	return NewTable(&df)
}

func (t *Table) ConcatColumn(tbl *Table) (*Table, error) {
	if t.DataFrame().Nrow() != tbl.DataFrame().Nrow() {
		return nil, fmt.Errorf("different number of row")
	}

	leftDf := *t.DataFrame()
	rightDf := *tbl.DataFrame()

	for _, col := range rightDf.Names() {
		leftDf = leftDf.Mutate(rightDf.Col(col))
	}

	t.dataFrame = &leftDf
	return t, nil
}
