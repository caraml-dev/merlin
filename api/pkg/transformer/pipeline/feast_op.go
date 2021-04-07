package pipeline

import "github.com/gojek/merlin/pkg/transformer/spec"

type FeastOp struct {
	featureTables []*spec.FeatureTable
}

func NewFeastOp(featureTables []*spec.FeatureTable) Op {
	return &FeastOp{
		featureTables: featureTables,
	}
}

func (f FeastOp) Execute(environment *Environment) error {
	panic("implement me")
}
