package scaler

import (
	mErrors "github.com/caraml-dev/merlin/pkg/errors"
	"github.com/caraml-dev/merlin/pkg/transformer/spec"
)

type Scaler interface {
	Validate() error
	Scale(values []interface{}) (interface{}, error)
}

func NewScaler(scalerSpec *spec.ScaleColumn) (Scaler, error) {
	var scalerImpl Scaler
	switch cfg := scalerSpec.ScalerConfig.(type) {
	case *spec.ScaleColumn_StandardScalerConfig:
		scalerImpl = &StandardScaler{cfg.StandardScalerConfig}
	case *spec.ScaleColumn_MinMaxScalerConfig:
		scalerImpl = &MinMaxScaler{cfg.MinMaxScalerConfig}
	default:
		return nil, mErrors.NewInvalidInputErrorf("scaler config has unexpected type %T", cfg)
	}
	return scalerImpl, nil
}
