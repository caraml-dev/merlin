package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/caraml-dev/merlin/pkg/inference-logger/mocks"
)

func Test(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		request        []byte
		requestHeader  map[string]string
		response       []byte
		responseHeader map[string]string
	}{
		{
			"nominal case",
			200,
			[]byte(`{"instances":[[0,0,0]]}`),
			map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			[]byte(`{"predictions":[[2]]}`),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			"nominal case with non json response",
			200,
			[]byte(`{"instances":[[0,0,0]]}`),
			map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			[]byte(`"predictions":[[2]]`),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			"nominal case with existing merlin log id",
			200,
			[]byte(`{"instances":[[0,0,0]]}`),
			map[string]string{
				"User-Agent":      "Mozilla/5.0",
				MerlinLogIdHeader: "4d93054e-5ad8-4e3c-9d97-a206560fc77b",
			},
			[]byte(`{"predictions":[[2]]}`),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			"error response with response body",
			500,
			[]byte(`{"instances":[[0,0,0]]}`),
			map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			[]byte(`{"error":"something went wrong}"`),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			"error response without response body",
			500,
			[]byte(`{"instances":[[0,0,0]]}`),
			map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			[]byte(``),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			"no request body",
			200,
			[]byte(``),
			map[string]string{
				"User-Agent": "Mozilla/5.0",
			},
			[]byte(`{"predictions":[[2]]}`),
			map[string]string{
				"Content-Type": "application/json",
			},
		},
	}
	logKinds := []LoggerSinkKind{NewRelic, Kafka}
	for _, logKind := range logKinds {
		for _, test := range tests {
			t.Run(fmt.Sprintf("%s-%s", logKind, test.name), func(t *testing.T) {
				// Start a predictor/model HTTP server
				predictor := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					merlinHeader := req.Header.Get(MerlinLogIdHeader)
					assert.True(t, IsValidUUID(merlinHeader), "header is not a valid UUID %s", merlinHeader)
					if len(test.requestHeader[MerlinLogIdHeader]) != 0 {
						assert.Equal(t, test.requestHeader[MerlinLogIdHeader], req.Header.Get(MerlinLogIdHeader))
					}

					// check that logger forward the request headers
					for k, v := range test.requestHeader {
						assert.Equal(t, v, req.Header.Get(k))
					}

					incomingRequest, err := io.ReadAll(req.Body)
					assert.Nil(t, err)

					for k, v := range test.responseHeader {
						rw.Header().Set(k, v)
					}

					rw.WriteHeader(test.statusCode)

					assert.Equal(t, test.request, incomingRequest)
					_, err = rw.Write(test.response)
					assert.Nil(t, err)

				}))
				// Close the server when test finishes
				defer predictor.Close()

				reader := bytes.NewReader(test.request)
				r := httptest.NewRequest("POST", "http://a", reader)
				for k, v := range test.requestHeader {
					r.Header.Set(k, v)
				}
				w := httptest.NewRecorder()

				targetUri, err := url.Parse(predictor.URL)
				assert.Nil(t, err)

				switch logKind {
				case NewRelic:
					l, _ := zap.NewDevelopment()
					zapLogger := l.Sugar()

					mockNewRelicLogsClient := &mocks.NewRelicLogsClient{}
					mockNewRelicLogsClient.On("CreateLogEntry", mock.Anything).Return(nil)

					workerConfig := &WorkerConfig{
						MinBatchSize: 1,
						MaxBatchSize: 5,
					}
					dispatcher := NewDispatcher(10, 100, workerConfig, logger, NewNewRelicSink(zapLogger, mockNewRelicLogsClient, serviceName, projectName, modelName, modelVersion), NewConsoleSink(logger))
					dispatcher.Start()
					httpProxy := httputil.NewSingleHostReverseProxy(targetUri)
					oh := NewLoggerHandler(dispatcher, LogModeAll, httpProxy, logger)

					oh.ServeHTTP(w, r)

					// wait for work to be pickedup
					time.Sleep(5 * time.Millisecond)

					respBody := w.Result().Body
					b2, _ := io.ReadAll(respBody)
					assert.Equal(t, test.response, b2)
					assert.Equal(t, test.statusCode, w.Code)
					// check that logger forward the response headers
					for k, v := range test.responseHeader {
						assert.Equal(t, v, w.Header().Get(k))
					}

					mockNewRelicLogsClient.AssertCalled(t, "CreateLogEntry", mock.Anything, mock.Anything)
					mockNewRelicLogsClient.AssertExpectations(t)

					_ = respBody.Close()
					dispatcher.Stop()
				case Kafka:
					l, _ := zap.NewDevelopment()
					zapLogger := l.Sugar()

					mockKafkaProducer := &mocks.KafkaProducer{}
					mockKafkaProducer.On("Produce", mock.Anything, mock.Anything).Return(nil)
					workerConfig := &WorkerConfig{
						MinBatchSize: 1,
						MaxBatchSize: 5,
					}
					dispatcher := NewDispatcher(10, 100, workerConfig, logger, NewKafkaSink(zapLogger, mockKafkaProducer, serviceName, projectName, modelName, modelVersion, topicName), NewConsoleSink(logger))
					dispatcher.Start()
					httpProxy := httputil.NewSingleHostReverseProxy(targetUri)
					oh := NewLoggerHandler(dispatcher, LogModeAll, httpProxy, logger)

					oh.ServeHTTP(w, r)

					// wait for work to be pickedup
					time.Sleep(5 * time.Millisecond)

					respBody := w.Result().Body
					b2, _ := io.ReadAll(respBody)
					assert.Equal(t, test.response, b2)
					assert.Equal(t, test.statusCode, w.Code)
					// check that logger forward the response headers
					for k, v := range test.responseHeader {
						assert.Equal(t, v, w.Header().Get(k))
					}

					mockKafkaProducer.AssertCalled(t, "Produce", mock.Anything, mock.Anything)
					mockKafkaProducer.AssertExpectations(t)

					_ = respBody.Close()
					dispatcher.Stop()
				}
			})
		}
	}
}

func IsValidUUID(uuid string) bool {
	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	return r.MatchString(uuid)
}
