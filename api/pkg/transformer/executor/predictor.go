package executor

import (
	"context"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	prt "github.com/gojek/merlin/pkg/protocol"
	"github.com/gojek/merlin/pkg/transformer/types"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/proto"
)

// ModelPredictor
type ModelPredictor interface {
	ModelPrediction(ctx context.Context, requestBody types.Payload, requestHeader map[string]string) (respBody types.Payload, respHeaders map[string]string, err error)
}

type mockModelPredictor struct {
	mockResponseBody   types.Payload
	mockResponseHeader map[string]string
	// responseConverterFunc is a function to convert incoming request to response that standard transformer accepts
	responseConverterFunc func(payload types.Payload) (types.Payload, error)
}

func newEchoMockPredictor() *mockModelPredictor {
	return &mockModelPredictor{}
}

// NewMockModelPredictor
func NewMockModelPredictor(respBody types.Payload, respHeader map[string]string, protocol prt.Protocol) *mockModelPredictor {
	var converterFn func(types.Payload) (types.Payload, error)
	if protocol == prt.UpiV1 {
		converterFn = upiResponseConverter
	} else {
		converterFn = restResponseConverter
	}

	return &mockModelPredictor{
		mockResponseBody:      respBody,
		mockResponseHeader:    respHeader,
		responseConverterFunc: converterFn,
	}
}

func restResponseConverter(payload types.Payload) (types.Payload, error) {
	return payload, nil
}

func upiResponseConverter(payload types.Payload) (types.Payload, error) {
	switch payloadT := payload.OriginalValue().(type) {
	case []byte:
		var resp upiv1.PredictValuesResponse
		if err := proto.Unmarshal(payloadT, &resp); err != nil {
			return nil, err
		}
		return (*types.UPIPredictionResponse)(&resp), nil
	case types.JSONObject:
		var resp upiv1.PredictValuesResponse
		if err := mapstructure.Decode(payloadT, &resp); err != nil {
			return nil, err
		}
		return (*types.UPIPredictionResponse)(&resp), nil
	case *upiv1.PredictValuesRequest:
		resp := &upiv1.PredictValuesResponse{}
		resp.PredictionResultTable = payloadT.PredictionTable
		return (*types.UPIPredictionResponse)(resp), nil
	case *upiv1.PredictValuesResponse:
		return payload, nil
	default:
		return nil, fmt.Errorf("unknown type of payload %T", payloadT)
	}
}

var _ ModelPredictor = (*mockModelPredictor)(nil)

// ModelPrediction
func (mock *mockModelPredictor) ModelPrediction(ctx context.Context, requestBody types.Payload, requestHeader map[string]string) (respBody types.Payload, respHeaders map[string]string, err error) {
	reqBodyObj, err := requestBody.AsInput()
	if err != nil {
		return nil, nil, err
	}

	respBody = reqBodyObj
	if mock.mockResponseBody != nil && !mock.mockResponseBody.IsNil() {
		respBody = mock.mockResponseBody
	}

	if mock.responseConverterFunc != nil {
		conversionRes, err := mock.responseConverterFunc(respBody)
		if err != nil {
			return nil, nil, err
		}
		respBody = conversionRes
	}

	respHeaders = requestHeader
	if mock.mockResponseHeader != nil {
		respHeaders = mock.mockResponseHeader
	}

	return respBody, respHeaders, nil
}
