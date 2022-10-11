package pipeline

import (
	"context"
	"fmt"

	"github.com/gojek/merlin/pkg/transformer/spec"
	"github.com/gojek/merlin/pkg/transformer/types"
	enc "github.com/gojek/merlin/pkg/transformer/types/encoder"
	"github.com/opentracing/opentracing-go"
)

type EncoderOp struct {
	encoderSpecs []*spec.Encoder
	*OperationTracing
}

type Encoder interface {
	Encode(values []interface{}, column string) (map[string]interface{}, error)
}

func NewEncoderOp(encoders []*spec.Encoder, tracingEnabled bool) *EncoderOp {
	encoderOp := &EncoderOp{encoderSpecs: encoders}
	if tracingEnabled {
		encoderOp.OperationTracing = NewOperationTracing(encoders, types.EncoderOpType)
	}
	return encoderOp
}

func (e *EncoderOp) Execute(ctx context.Context, env *Environment) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "pipeline.EncoderOp")
	defer span.Finish()

	var encoderImpl Encoder
	for _, encoderSpec := range e.encoderSpecs {
		switch encoderCfg := encoderSpec.EncoderConfig.(type) {
		case *spec.Encoder_OrdinalEncoderConfig:
			ordinalEncoder, err := enc.NewOrdinalEncoder(encoderCfg.OrdinalEncoderConfig)
			if err != nil {
				return err
			}
			encoderImpl = ordinalEncoder
		case *spec.Encoder_CyclicalEncoderConfig:
			cyclicalEncoder, err := enc.NewCyclicalEncoder(encoderCfg.CyclicalEncoderConfig)
			if err != nil {
				return err
			}
			encoderImpl = cyclicalEncoder
		default:
			return fmt.Errorf("encoder spec have unexpected type %T", encoderCfg)
		}
		env.SetSymbol(encoderSpec.Name, encoderImpl)
		if e.OperationTracing != nil {
			if err := e.AddInputOutput(nil, map[string]interface{}{encoderSpec.Name: "The result of this operation is on the transformer step that use this encoder"}); err != nil {
				return err
			}
		}
	}
	return nil
}
