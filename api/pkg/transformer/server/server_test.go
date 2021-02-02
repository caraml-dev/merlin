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
	tests := []struct {
		name            string
		request         []byte
		requestHeader   map[string]string
		expModelRequest []byte
		modelResponse   []byte
		modelStatusCode int
	}{
		{
			"nominal case",
			[]byte(`{"predictions": [2, 2]}`),
			map[string]string{MerlinLogIdHeader: "1234"},
			[]byte(`{"driver_id":"1001","preprocess":true}`),
			[]byte(`{"predictions": [2, 2]}`),
			200,
		},
		{
			"model return error",
			[]byte(`{"predictions": [2, 2]}`),
			map[string]string{MerlinLogIdHeader: "1234"},
			[]byte(`{"driver_id":"1001","preprocess":true}`),
			[]byte(`{"predictions": [2, 2]}`),
			500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockPreprocessHandler := func(ctx context.Context, request []byte) ([]byte, error) {
				return test.expModelRequest, nil
			}

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				// check log id is propagated as header
				logId, ok := test.requestHeader[MerlinLogIdHeader]
				if ok {
					assert.Equal(t, logId, r.Header.Get(MerlinLogIdHeader))
				}

				assert.Nil(t, err)
				assert.Equal(t, test.expModelRequest, body)
				w.WriteHeader(test.modelStatusCode)
				w.Write(test.modelResponse)
			}))
			defer ts.Close()

			rr := httptest.NewRecorder()

			reqBody := bytes.NewBuffer(test.request)
			req, err := http.NewRequest("POST", ts.URL, reqBody)
			for k, v := range test.requestHeader {
				req.Header.Set(k, v)
			}

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
			assert.Equal(t, test.modelResponse, response)
			assert.Equal(t, test.modelStatusCode, rr.Code)
		})
	}
}
