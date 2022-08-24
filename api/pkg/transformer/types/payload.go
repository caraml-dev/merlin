package types

import (
	"encoding/json"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

type UPIPredictionRequest upiv1.PredictValuesRequest
type UPIPredictionResponse upiv1.PredictValuesResponse
type BytePayload []byte

type Payload interface {
	AsInput() (Payload, error)
	AsOutput() (Payload, error)
	IsNil() bool
}

func (upr *UPIPredictionRequest) AsInput() (Payload, error) {
	return upr, nil
}

func (upr *UPIPredictionRequest) AsOutput() (Payload, error) {
	return upr, nil
}

func (upr *UPIPredictionRequest) IsNil() bool {
	return upr == nil
}

func (upr *UPIPredictionResponse) AsInput() (Payload, error) {
	return upr, nil
}

func (upr *UPIPredictionResponse) AsOutput() (Payload, error) {
	return upr, nil
}

func (upr *UPIPredictionResponse) IsNil() bool {
	return upr == nil
}

func (bp BytePayload) AsInput() (Payload, error) {
	var obj JSONObject
	if err := json.Unmarshal(bp, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

func (bp BytePayload) AsOutput() (Payload, error) {
	return bp, nil
}

func (bp BytePayload) IsNil() bool {
	return bp == nil
}

func (jo JSONObject) AsInput() (Payload, error) {
	return jo, nil
}

func (jo JSONObject) AsOutput() (Payload, error) {
	marshalOut, err := json.Marshal(jo)
	if err != nil {
		return nil, err
	}
	return BytePayload(marshalOut), nil
}

func (jo JSONObject) IsNil() bool {
	return jo == nil
}
