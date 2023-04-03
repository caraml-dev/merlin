package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	kafkalib "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Producer responsible to produce log to kafka
type Producer struct {
	topic               string
	producer            producer
	serializationFormat SerializationFormat
	logger              *zap.Logger
}

type producer interface {
	Produce(*kafkalib.Message, chan kafkalib.Event) error
	Close()
	Events() chan kafkalib.Event
}

// NewProducer creates Producer instance
func NewProducer(cfg Config, logger *zap.Logger) (*Producer, error) {
	producer, err := kafkalib.NewProducer(&kafkalib.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
		"message.max.bytes": cfg.MaxMessageSizeBytes,
		"compression.type":  cfg.CompressionType,
	})
	if err != nil {
		return nil, err
	}

	// Test that we are able to query the broker on the topic. If the topic
	// does not already exist on the broker, this should create it.
	_, err = producer.GetMetadata(&cfg.Topic, false, cfg.ConnectTimeoutMS)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Error Querying topic %s from Kafka broker(s)", cfg.Topic)
	}

	kafkaProducer := &Producer{
		topic:               cfg.Topic,
		producer:            producer,
		serializationFormat: cfg.SerializationFmt,
		logger:              logger,
	}

	go kafkaProducer.handleMessageDelivery()
	return kafkaProducer, nil
}

// Close a producer instance
func (kp *Producer) Close() {
	kp.producer.Close()
}

// Produce log by publishing it to kafka
func (kp *Producer) Produce(ctx context.Context, value interface{}) error {
	var msgValue []byte
	if kp.serializationFormat == JSON {
		val, err := json.Marshal(value)
		if err != nil {
			return err
		}
		msgValue = val
	} else if kp.serializationFormat == Protobuf {
		protoVal, validProtoType := value.(proto.Message)
		if !validProtoType {
			return fmt.Errorf("can't serialize if type is not proto.Message")
		}
		val, err := proto.Marshal(protoVal)
		if err != nil {
			return err
		}
		msgValue = val
	}
	err := kp.producer.Produce(&kafkalib.Message{
		TopicPartition: kafkalib.TopicPartition{
			Topic:     &kp.topic,
			Partition: kafkalib.PartitionAny,
		},
		Value: msgValue,
		Key:   nil,
	}, nil)

	return err
}

func (k *Producer) handleMessageDelivery() {
	for e := range k.producer.Events() {
		switch ev := e.(type) {
		case *kafkalib.Message:
			if ev.TopicPartition.Error != nil {
				val := k.deserializeMessageValue(ev.Value)
				k.logger.Warn("failed to deliver message", zap.Error(ev.TopicPartition.Error), zap.Any("value", val))
			}
		}
	}
}

func (k *Producer) deserializeMessageValue(value []byte) any {
	if k.serializationFormat == JSON {
		var output map[string]any
		err := json.Unmarshal(value, &output)
		if err != nil {
			k.logger.Warn("failed to deserialize json message value", zap.Error(err))
		}
		return output
	}
	// for Protobuf case
	predictionLog := &upiv1.PredictionLog{}
	if err := proto.Unmarshal(value, predictionLog); err != nil {
		k.logger.Warn("failed to deserialize protobuf message value", zap.Error(err))
	}
	return predictionLog
}
