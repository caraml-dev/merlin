package pipeline

import "context"

// PredictionLog
type PredictionLogProducer interface {
	Produce(ctx context.Context, value interface{}) error
}
