package kafka

import (
	"context"
	"fmt"
	"testing"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// mockKafkaProducer implements the kafkaProducer
type mockKafkaProducer struct {
	mock.Mock
}

func (mp *mockKafkaProducer) GetMetadata(
	topic *string,
	allTopics bool,
	timeoutMs int,
) (*kafka.Metadata, error) {
	mp.Called(topic, allTopics, timeoutMs)
	return nil, nil
}

func (mp *mockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	ret := mp.Called(msg, deliveryChan)
	// Send event to deliveryChan

	return ret.Error(0)
}

func (mp *mockKafkaProducer) Close() {
	mp.Called()
}

func (mp *mockKafkaProducer) Events() chan kafka.Event {
	ret := mp.Called()
	var r0 chan kafka.Event
	if ret.Get(0) != nil {
		return ret.Get(0).(chan kafka.Event)
	}
	return r0
}

func TestProducer_Produce(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	topic := "topic"
	type fields struct {
		topic               string
		producer            func(predictionLog *upiv1.PredictionLog) *mockKafkaProducer
		serializationFormat SerializationFormat
		logger              *zap.Logger
	}
	type args struct {
		value *upiv1.PredictionLog
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				topic: "topic",
				producer: func(predictionLog *upiv1.PredictionLog) *mockKafkaProducer {
					mockProducer := &mockKafkaProducer{}
					val, _ := proto.Marshal(predictionLog)
					mockProducer.On("Produce", &kafka.Message{
						TopicPartition: kafka.TopicPartition{
							Topic:     &topic,
							Partition: kafka.PartitionAny,
						},
						Value: val,
						Key:   nil,
					}, mock.Anything).Return(nil)
					return mockProducer
				},
				serializationFormat: Protobuf,
				logger:              logger,
			},
			args: args{
				value: &upiv1.PredictionLog{
					PredictionId: "predictionID",
					ProjectName:  "project",
					ModelName:    "model",
					ModelVersion: "version",
					TargetName:   "target",
				},
			},
			wantErr: false,
		},
		{
			name: "failed",
			fields: fields{
				topic: "topic",
				producer: func(predictionLog *upiv1.PredictionLog) *mockKafkaProducer {
					mockProducer := &mockKafkaProducer{}
					val, _ := proto.Marshal(predictionLog)
					mockProducer.On("Produce", &kafka.Message{
						TopicPartition: kafka.TopicPartition{
							Topic:     &topic,
							Partition: kafka.PartitionAny,
						},
						Value: val,
						Key:   nil,
					}, mock.Anything).Return(fmt.Errorf("kafka broker down"))
					return mockProducer
				},
				serializationFormat: Protobuf,
				logger:              logger,
			},
			args: args{
				value: &upiv1.PredictionLog{
					PredictionId: "predictionID",
					ProjectName:  "project",
					ModelName:    "model",
					ModelVersion: "version",
					TargetName:   "target",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kProducer := tt.fields.producer(tt.args.value)
			kp := &Producer{
				topic:               tt.fields.topic,
				producer:            kProducer,
				serializationFormat: tt.fields.serializationFormat,
				logger:              tt.fields.logger,
			}
			if err := kp.Produce(context.Background(), tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Producer.Produce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
