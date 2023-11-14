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
                        prediction_columns=["accept", "label"],
                        prediction_column_types=[ValueType.BOOLEAN, ValueType.STRING],
                        prediction_label_column="label",
                        prediction_score_column="accept",
                        timestamp_column="request_timestamp",
                        feature_columns=["distance", "transaction"],
                        feature_column_types=[ValueType.INT64, ValueType.FLOAT64],
                    ),
                ),
                observability_backend=ObservabilityBackend(
                    type=ObservabilityBackendType.ARIZE,
                    arize_config=ArizeConfig(
                        api_key="SECRET_API_KEY",
                        space_key="SECRET_SPACE_KEY",
                    ),
                ),
                kafka=KafkaConsumerConfig(
                    topic="test-topic",
                    bootstrap_servers="localhost:9092",
                    group_id="test-group",
                    auto_offset_reset="latest",
                    batch_size=100,
                    poll_timeout_seconds=1.0,
                ),
            )
        )
        assert cfg.environment.model == dataclasses.asdict(
            expected_cfg.environment.model
        )
        assert cfg.environment.observability_backend == dataclasses.asdict(
            expected_cfg.environment.observability_backend
        )
        assert cfg.environment.kafka == dataclasses.asdict(
            expected_cfg.environment.kafka
        )
