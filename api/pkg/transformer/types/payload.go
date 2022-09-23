package types

import (
	"encoding/json"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
)

// UPIPredictionRequest, wrapper of upiv1.PredictValuesRequest
type UPIPredictionRequest upiv1.PredictValuesRequest

// UPIPredictionResponse, wrapper of upiv1.PredictValuesResponse
type UPIPredictionResponse upiv1.PredictValuesResponse

// BytePayload, wrapper of []byte
type BytePayload []byte

// Payload represent all accepted payload for standard transformer
type Payload interface {
	// AsInput convert current Payload into format that accepted by standard transformer input
	// Supported input in standard transformer
	// 1. JSONObject or map[string]interface{}. This type is used by HTTP_JSON protocol
	// 2. UPIPredictionRequest. This type is used by UPI_V1 protocol and in preprocessing step
	// 3. UPIPredictionResponse. This type is used by UP1_V1 protocol and in postprocessing step
	AsInput() (Payload, error)

	// AsOutput convert current Payload into format that later on may be sent through network
	// Supported output type
	// 1. BytePayload. This type is used by HTTP_JSON
	// 2. UPIPredictionRequest. This type is used by UPI_V1 protocol
	// 3. UPIPredictionResponse. This type is used by UPI_V1 protocol
	AsOutput() (Payload, error)

	// IsNil check whether the interface is nil
	IsNil() bool

	// OriginalValue will unwrap the Payload to its origin type
	OriginalValue() any
}

// AsInput convert current Payload into format that accepted by standard transformer input
func (upr *UPIPredictionRequest) AsInput() (Payload, error) {
	return upr, nil
}

// AsOutput convert current Payload into format that later on may be sent through network
func (upr *UPIPredictionRequest) AsOutput() (Payload, error) {
	return upr, nil
}

// IsNil check whether the interface is nil
func (upr *UPIPredictionRequest) IsNil() bool {
	return upr == nil
}

// OriginalValue will unwrap the Payload to its origin type
func (upr *UPIPredictionRequest) OriginalValue() any {
	return (*upiv1.PredictValuesRequest)(upr)
}

// AsInput convert current Payload into format that accepted by standard transformer input
func (upr *UPIPredictionResponse) AsInput() (Payload, error) {
	return upr, nil
}

// AsOutput convert current Payload into format that later on may be sent through network
func (upr *UPIPredictionResponse) AsOutput() (Payload, error) {
	return upr, nil
}

// IsNil check whether the interface is nil
func (upr *UPIPredictionResponse) IsNil() bool {
	return upr == nil
}

// OriginalValue will unwrap the Payload to its origin type
func (upr *UPIPredictionResponse) OriginalValue() any {
	return (*upiv1.PredictValuesResponse)(upr)
}

// AsInput convert current Payload into format that accepted by standard transformer input
func (bp BytePayload) AsInput() (Payload, error) {
	var obj JSONObject
	if err := json.Unmarshal(bp, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// AsOutput convert current Payload into format that later on may be sent through network
func (bp BytePayload) AsOutput() (Payload, error) {
	return bp, nil
}

// IsNil check whether the interface is nil
func (bp BytePayload) IsNil() bool {
	return bp == nil
}

// OriginalValue will unwrap the Payload to its origin type
func (bp BytePayload) OriginalValue() any {
	return ([]byte)(bp)
}

// AsInput convert current Payload into format that accepted by standard transformer input
func (jo JSONObject) AsInput() (Payload, error) {
	return jo, nil
}

// AsOutput convert current Payload into format that later on may be sent through network
func (jo JSONObject) AsOutput() (Payload, error) {
	marshalOut, err := json.Marshal(jo)
	if err != nil {
		return nil, err
	}
	return BytePayload(marshalOut), nil
}

// IsNil check whether the interface is nil
func (jo JSONObject) IsNil() bool {
	return jo == nil
}

// OriginalValue will unwrap the Payload to its origin type
func (jo JSONObject) OriginalValue() any {
	return jo
}
