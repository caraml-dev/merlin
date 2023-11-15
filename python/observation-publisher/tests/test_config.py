import dataclasses

from hydra import compose, initialize

from publisher.config import (
    ArizeConfig,
    Environment,
    KafkaConsumerConfig,
    ModelSchema,
    ModelSpec,
    ModelType,
    ObservabilityBackend,
    ObservabilityBackendType,
    PublisherConfig,
    ValueType,
    PredictionLogConsumerConfig,
    ConsumerType,
)


def test_config_initialization():
    with initialize(version_base=None, config_path="../conf"):
        cfg = compose(config_name="config", overrides=["+environment=example-override"])
        expected_cfg = PublisherConfig(
            environment=Environment(
                model=ModelSpec(
                    id="test-model",
                    version="0.1.0",
                    type=ModelType.BINARY_CLASSIFICATION,
                    schema=ModelSchema(
                        column_types={
                            "accept": ValueType.BOOLEAN,
                            "label": ValueType.STRING,
                            "distance": ValueType.INT64,
                            "transaction": ValueType.FLOAT64,
                        },
                        prediction_label_column="label",
                        prediction_score_column="accept",
                        timestamp_column="request_timestamp",
                    ),
                ),
                observability_backend=ObservabilityBackend(
                    type=ObservabilityBackendType.ARIZE,
                    arize_config=ArizeConfig(
                        api_key="SECRET_API_KEY",
                        space_key="SECRET_SPACE_KEY",
                    ),
                ),
                consumer=PredictionLogConsumerConfig(
                    type=ConsumerType.KAFKA,
                    kafka_config=KafkaConsumerConfig(
                        topic="test-topic",
                        bootstrap_servers="localhost:9092",
                        group_id="test-group",
                        poll_timeout_seconds=1.0,
                        additional_consumer_config={
                            "auto.offset.reset": "latest",
                        },
                    ),
                ),
            )
        )
        assert cfg.environment.model == dataclasses.asdict(
            expected_cfg.environment.model
        )
        assert cfg.environment.observability_backend == dataclasses.asdict(
            expected_cfg.environment.observability_backend
        )
        assert cfg.environment.consumer == dataclasses.asdict(
            expected_cfg.environment.consumer
        )
