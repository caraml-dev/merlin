package deployment

import (
	"fmt"
	"strings"

	"github.com/caraml-dev/merlin/models"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/mapstructure"
)

type ConsumerConfig struct {
	ModelID          string             `yaml:"model_id"`
	ModelVersion     string             `yaml:"model_version"`
	InferenceSchema  *models.SchemaSpec `yaml:"inference_schema"`
	ObservationSinks []ObservationSink  `yaml:"observation_sinks"`
}

type KafkaSource struct {
	Topic                    string            `yaml:"topic"`
	BootstrapServers         string            `yaml:"bootstrap_servers"`
	GroupID                  string            `yaml:"group_id"`
	BatchSize                int               `yaml:"batch_size"`
	AdditionalConsumerConfig map[string]string `yaml:"additional_consumer_config"`
}

type SinkType string

const (
	Arize SinkType = "ARIZE"
	BQ    SinkType = "BIGQUERY"

	PublisherRevisionAnnotationKey = "publisher-revision"
)

type ObservationSink struct {
	Type   SinkType `yaml:"type"`
	Config any      `yaml:"config"`
}

type ArizeSink struct {
	APIKey   string `yaml:"api_key"`
	SpaceKey string `yaml:"space_key"`
}

type BigQuerySink struct {
	Project string `yaml:"project"`
	Dataset string `yaml:"dataset"`
	TTLDays int    `yaml:"ttl_days"`
}

func (os *ObservationSink) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var observationSinkRaw map[string]any

	if err := unmarshal(&observationSinkRaw); err != nil {
		return err
	}

	sinkType, ok := observationSinkRaw["type"].(string)
	if !ok {
		return fmt.Errorf("sink type is not correct %T", observationSinkRaw["type"])
	}

	os.Type = SinkType(sinkType)
	if os.Type == Arize {
		var arizeConfig ArizeSink
		cfg, err := decodeMap[*ArizeSink](observationSinkRaw["config"], &arizeConfig)
		if err != nil {
			return err
		}
		os.Config = cfg
	} else if os.Type == BQ {
		var bqConfig BigQuerySink
		cfg, err := decodeMap[*BigQuerySink](observationSinkRaw["config"], &bqConfig)
		if err != nil {
			return err
		}
		os.Config = cfg
	} else {
		return fmt.Errorf(`type '%s' is not supported`, os.Type)
	}
	return nil
}

func decodeMap[V *ArizeSink | *BigQuerySink](raw any, result V) (V, error) {
	decoderCfg := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           result,
		MatchName: func(mapKey, fieldName string) bool {
			mapKeySnakeCase := strcase.ToCamel(mapKey)
			fieldNameSnakeCase := strcase.ToCamel(fieldName)
			return strings.EqualFold(mapKeySnakeCase, fieldNameSnakeCase)
		},
	}
	decoder, err := mapstructure.NewDecoder(decoderCfg)
	if err != nil {
		return nil, err
	}
	if err := decoder.Decode(raw); err != nil {
		return nil, err
	}
	return result, nil
}
