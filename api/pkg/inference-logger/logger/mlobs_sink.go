package logger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	ErrMalformedLogEntry = errors.New("malformed log entry")
)

type MLObsSink struct {
	logger   *zap.SugaredLogger
	producer KafkaProducer

	projectName  string
	modelName    string
	modelVersion string
}

type StandardModelRequest struct {
	Instances [][]*float64 `json:"instances"`
	SessionId string       `json:"session_id"`
	RowIds    []string     `json:"row_ids"`
}

type StandardModelResponse struct {
	Predictions []float64 `json:"predictions"`
}

func NewMLObsSink(
	logger *zap.SugaredLogger,
	producer KafkaProducer,
	adminClient KafkaAdmin,
	projectName string,
	modelName string,
	modelVersion string,
) (LogSink, error) {
	sink := &MLObsSink{
		logger:       logger,
		producer:     producer,
		projectName:  projectName,
		modelName:    modelName,
		modelVersion: modelVersion,
	}
	topicResults, err := adminClient.CreateTopics(context.Background(), []kafka.TopicSpecification{
		{
			Topic:             sink.topicName(),
			NumPartitions:     NumPartitions,
			ReplicationFactor: ReplicationFactor,
		},
	})
	for _, result := range topicResults {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			return nil, err
		}
	}
	go sink.handleMessageDelivery()

	return sink, nil
}

func (m *MLObsSink) newPredictionLog(rawLogEntry *LogEntry) (*upiv1.PredictionLog, error) {
	predictionLog := &upiv1.PredictionLog{}
	standardModelRequest := &StandardModelRequest{}
	err := json.Unmarshal(rawLogEntry.RequestPayload.Body, standardModelRequest)
	if err != nil {
		return nil, err
	}
	standardModelResponse := &StandardModelResponse{}
	err = json.Unmarshal(rawLogEntry.ResponsePayload.Body, standardModelResponse)
	if err != nil {
		return nil, err
	}
	// If there is only one prediction, session id alone is enough to for unique prediction id
	if len(standardModelRequest.Instances) == 1 && standardModelRequest.RowIds == nil {
		standardModelRequest.RowIds = []string{""}
	}

	if len(standardModelRequest.RowIds) != len(standardModelResponse.Predictions) {
		return nil, fmt.Errorf("%w: number of row ids and predictions do not match", ErrMalformedLogEntry)
	}
	if standardModelRequest.SessionId == "" {
		return nil, fmt.Errorf("%w: missing session id", ErrMalformedLogEntry)
	}
	predictionLog.PredictionId = standardModelRequest.SessionId
	predictionLog.ModelName = m.modelName
	predictionLog.ModelVersion = m.modelVersion
	predictionLog.ProjectName = m.projectName
	featuresTable := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	predictionTable := &structpb.Struct{
		Fields: make(map[string]*structpb.Value),
	}
	featureTableData := &structpb.ListValue{
		Values: make([]*structpb.Value, len(standardModelRequest.Instances)),
	}
	predictionTableData := &structpb.ListValue{
		Values: make([]*structpb.Value, len(standardModelResponse.Predictions)),
	}
	rowIds := &structpb.ListValue{
		Values: make([]*structpb.Value, len(standardModelRequest.RowIds)),
	}
	for i, instance := range standardModelRequest.Instances {
		featureTableRow := &structpb.ListValue{
			Values: make([]*structpb.Value, len(instance)),
		}
		for j, value := range instance {
			if value != nil {
				featureTableRow.Values[j] = structpb.NewNumberValue(*value)
			} else {
				featureTableRow.Values[j] = structpb.NewNullValue()
			}
		}

		featureTableData.Values[i] = structpb.NewListValue(featureTableRow)
		predictionTableData.Values[i] = structpb.NewNumberValue(standardModelResponse.Predictions[i])
		rowIds.Values[i] = structpb.NewStringValue(standardModelRequest.RowIds[i])
	}
	featuresTable.Fields["data"] = structpb.NewListValue(featureTableData)
	featuresTable.Fields["row_ids"] = structpb.NewListValue(rowIds)
	predictionTable.Fields["data"] = structpb.NewListValue(predictionTableData)
	predictionTable.Fields["row_ids"] = structpb.NewListValue(rowIds)

	predictionLog.Input = &upiv1.ModelInput{
		FeaturesTable: featuresTable,
	}
	predictionLog.Output = &upiv1.ModelOutput{
		PredictionResultsTable: predictionTable,
	}
	return predictionLog, nil
}

func (m *MLObsSink) topicName() string {
	return fmt.Sprintf("caraml-%s-%s-%s-prediction-log", m.projectName, m.modelName, m.modelVersion)
}

func (m *MLObsSink) buildNewKafkaMessage(predictionLog *upiv1.PredictionLog) (*kafka.Message, error) {
	logBytes, err := proto.Marshal(predictionLog)
	if err != nil {
		return nil, err
	}
	topic := fmt.Sprintf("caraml-%s-%s-%s-prediction-log", m.projectName, m.modelName, m.modelVersion)
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
		Value: logBytes,
	}, nil
}

func (m *MLObsSink) Sink(rawLogEntries []*LogEntry) error {
	for _, rawLogEntry := range rawLogEntries {
		predictionLog, err := m.newPredictionLog(rawLogEntry)
		if err != nil {
			m.logger.Errorf("unable to convert log entry: %v", err)
		}
		kafkaMessage, err := m.buildNewKafkaMessage(predictionLog)
		if err != nil {
			m.logger.Errorf("unable to build kafka message: %v", err)
		}
		err = m.producer.Produce(kafkaMessage, m.producer.Events())
		if err != nil {
			m.logger.Errorf("unable to produce kafka message: %v", err)
		}
	}
	return nil
}

func (m *MLObsSink) handleMessageDelivery() {
	for e := range m.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				m.logger.Errorf("Delivery failed: %v\n", ev.TopicPartition.Error)
			}
		}
	}
}
