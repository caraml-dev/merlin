package executor

import (
	"context"

	"github.com/gojek/merlin/pkg/transformer/types"
)

// ModelPredictor
type ModelPredictor interface {
	ModelPrediction(ctx context.Context, requestBody types.Payload, requestHeader map[string]string) (respBody types.Payload, respHeaders map[string]string, err error)
}

type mockModelPredictor struct {
	mockResponseBody   types.Payload
	mockResponseHeader map[string]string
}

func newEchoMockPredictor() *mockModelPredictor {
	return &mockModelPredictor{}
}

func NewMockModelPredictor(respBody types.Payload, respHeader map[string]string) *mockModelPredictor {
	return &mockModelPredictor{
		mockResponseBody:   respBody,
		mockResponseHeader: respHeader,
	}
}

var _ ModelPredictor = (*mockModelPredictor)(nil)

func (mock *mockModelPredictor) ModelPrediction(ctx context.Context, requestBody types.Payload, requestHeader map[string]string) (respBody types.Payload, respHeaders map[string]string, err error) {
	reqBodyObj, err := requestBody.AsInput()
	if err != nil {
		return nil, nil, err
	}

	respBody = reqBodyObj
	if mock.mockResponseBody != nil && !mock.mockResponseBody.IsNil() {
		respBody = mock.mockResponseBody
	}

	respHeaders = requestHeader
	if mock.mockResponseHeader != nil {
		respHeaders = mock.mockResponseHeader
	}

	return respBody, respHeaders, nil
}
