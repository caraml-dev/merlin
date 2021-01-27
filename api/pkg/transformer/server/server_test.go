package server

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_PredictHandler_NoTransformation(t *testing.T) {
	mockPredictResponse := []byte(`{"predictions": [2, 2]}`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(mockPredictResponse)
	}))
	defer ts.Close()

	rr := httptest.NewRecorder()

	reqBody := bytes.NewBufferString(`{"driver_id":"1001"}`)
	req, err := http.NewRequest("POST", ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		ModelPredictURL: ts.URL,
	}
	logger, _ := zap.NewDevelopment()
	server := New(options, logger)

	server.PredictHandler(rr, req)

	response, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	assert.Equal(t, mockPredictResponse, response)
}

func TestServer_PredictHandler_WithPreprocess(t *testing.T) {
	mockPreprocessResponse := []byte(`{"driver_id":"1001","preprocess":true}`)
	mockPreprocessHandler := func(ctx context.Context, request []byte) ([]byte, error) {
		return mockPreprocessResponse, nil
	}

	mockPredictResponse := []byte(`{"predictions": [2, 2]}`)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, mockPreprocessResponse, body)
		w.Write(mockPredictResponse)
	}))
	defer ts.Close()

	rr := httptest.NewRecorder()

	reqBody := bytes.NewBufferString(`{"driver_id":"1001"}`)
	req, err := http.NewRequest("POST", ts.URL, reqBody)
	if err != nil {
		t.Fatal(err)
	}

	options := &Options{
		ModelPredictURL: ts.URL,
	}
	logger, _ := zap.NewDevelopment()
	server := New(options, logger)
	server.PreprocessHandler = mockPreprocessHandler

	server.PredictHandler(rr, req)

	response, err := ioutil.ReadAll(rr.Body)
	assert.Nil(t, err)
	assert.Equal(t, mockPredictResponse, response)
}
