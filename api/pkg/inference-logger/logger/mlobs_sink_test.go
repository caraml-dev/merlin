package logger

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
)

func asInstanceValue(value float64) *float64 {
	return &value
}

func newTestStandardModelRequest(data [][]*float64) *StandardModelRequest {
	rowIds := make([]string, len(data))
	for i := 0; i < len(data); i++ {
		rowIds[i] = uuid.New().String()
	}
	return &StandardModelRequest{
		Instances: data,
		SessionId: uuid.New().String(),
		RowIds:    rowIds,
	}
}

func newTestStandardModelResponse(data []float64) *StandardModelResponse {
	return &StandardModelResponse{
		Predictions: data,
	}
}

func newTestLogEntry(request *StandardModelRequest, response *StandardModelResponse) *LogEntry {
	requestBodyBytes, _ := json.Marshal(request)
	requestPayload := &RequestPayload{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: requestBodyBytes,
	}
	responseBodyBytes, _ := json.Marshal(response)
	responsePayload := &ResponsePayload{
		StatusCode: 200,
		Body:       responseBodyBytes,
	}
	return &LogEntry{
		RequestId:       uuid.New().String(),
		EventTimestamp:  nil,
		RequestPayload:  requestPayload,
		ResponsePayload: responsePayload,
	}
}

func TestLogEntryToPredictionLogConversion(t *testing.T) {
	tests := []struct {
		name     string
		request  *StandardModelRequest
		response *StandardModelResponse
	}{
		{
			"multi rows predictions",
			newTestStandardModelRequest([][]*float64{
				{asInstanceValue(1.0), asInstanceValue(2.0)},
				{asInstanceValue(3.0), asInstanceValue(4.0)},
			}),
			newTestStandardModelResponse([]float64{0.5, 0.7}),
		},
		{
			"null feature values",
			newTestStandardModelRequest([][]*float64{
				{nil},
			}),
			newTestStandardModelResponse([]float64{0.0}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := &MLObsSink{
				modelName:    "test-model",
				modelVersion: "1",
				projectName:  "test-project",
			}
			logEntry := newTestLogEntry(tt.request, tt.response)
			predictionLog, err := sink.newPredictionLog(logEntry)
			if err != nil {
				t.Errorf("unable to convert log entry: %v", err)
			}
			for i, row := range predictionLog.
				GetInput().
				GetFeaturesTable().
				GetFields()["data"].GetListValue().GetValues() {
				for j, col := range row.GetListValue().GetValues() {
					switch col.GetKind().(type) {
					case *structpb.Value_NullValue:
						if tt.request.Instances[i][j] != nil {
							t.Errorf("feature value should have been nil on index (%d, %d)", i, j)
						}
					default:
						if col.GetNumberValue() != *tt.request.Instances[i][j] {
							t.Errorf("unexpected feature data on index (%d, %d)", i, j)
						}
					}
				}
			}

			if predictionLog.ModelVersion != sink.modelVersion {
				t.Errorf("unexpected model version: %s", predictionLog.ModelVersion)
			}
			if predictionLog.ModelName != sink.modelName {
				t.Errorf("unexpected model name: %s", predictionLog.ModelName)
			}
			if predictionLog.ProjectName != sink.projectName {
				t.Errorf("unexpected project name: %s", predictionLog.ProjectName)
			}
		})
	}
}
