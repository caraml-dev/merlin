package series

import "github.com/go-gota/gota/series"

type Type string

const (
	String Type = "string"
	Int    Type = "int"
	Float  Type = "float"
	Bool   Type = "bool"
)

type Series struct {
	series *series.Series
}

func NewSeries(s *series.Series) *Series {
	return &Series{s}
}

func New(values interface{}, t Type, name string) *Series {
	s := series.New(values, series.Type(t), name)
	return &Series{&s}
}

func (s *Series) Series() *series.Series {
	return s.series
}

func (s *Series) Type() Type {
	return Type(s.series.Type())
}
