package executor

import (
	"context"
	"encoding/json"

	"github.com/gojek/merlin/pkg/transformer/types"
)

// ModelPredictor
type ModelPredictor interface {
	ModelPrediction(ctx context.Context, requestBody []byte, requestHeader map[string]string) (respBody types.JSONObject, respHeaders map[string]string, err error)
}

type mockModelPredictor struct {
	mockResponseBody   types.JSONObject
	mockResponseHeader map[string]string
}

func newEchoMockPredictor() *mockModelPredictor {
	return &mockModelPredictor{}
}

func NewMockPredictor(respBody types.JSONObject, respHeader map[string]string) *mockModelPredictor {
	return &mockModelPredictor{
		mockResponseBody:   respBody,
		mockResponseHeader: respHeader,
	}
}

var _ ModelPredictor = (*mockModelPredictor)(nil)

func (mock *mockModelPredictor) ModelPrediction(ctx context.Context, requestBody []byte, requestHeader map[string]string) (respBody types.JSONObject, respHeaders map[string]string, err error) {
	var reqBodyObj types.JSONObject
	if err := json.Unmarshal(requestBody, &reqBodyObj); err != nil {
		return nil, nil, err
	}
	respBody = reqBodyObj
	if mock.mockResponseBody != nil {
		respBody = mock.mockResponseBody
	}
	respHeaders = requestHeader
	if mock.mockResponseHeader != nil {
		respHeaders = mock.mockResponseHeader
	}
	return respBody, respHeaders, nil
}
