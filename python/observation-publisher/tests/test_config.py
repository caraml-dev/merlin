import dataclasses

from hydra import compose, initialize
from merlin.observability.inference import (BinaryClassificationOutput,
                                            InferenceSchema, InferenceType,
                                            ValueType)

from publisher.config import (ArizeConfig, Environment, KafkaConsumerConfig,
                              ObservabilityBackend, ObservabilityBackendType,
                              ObservationSource, ObservationSourceConfig,
                              PublisherConfig)


def test_config_initialization():
    with initialize(version_base=None, config_path="../conf"):
        cfg = compose(config_name="config", overrides=["+environment=example-override"])
        expected_cfg = PublisherConfig(
            environment=Environment(
                model_id="test-model",
                model_version="0.1.0",
                inference_schema=InferenceSchema(
                    type=InferenceType.BINARY_CLASSIFICATION,
                    feature_types={
                        "distance": ValueType.INT64,
                        "transaction": ValueType.FLOAT64,
                    },
                    binary_classification=BinaryClassificationOutput(
                        prediction_label_column="label",
                        prediction_score_column="score",
                    ),
                ),
                observability_backend=ObservabilityBackend(
                    type=ObservabilityBackendType.ARIZE,
                    arize_config=ArizeConfig(
                        api_key="SECRET_API_KEY",
                        space_key="SECRET_SPACE_KEY",
                    ),
                ),
                observation_source=ObservationSourceConfig(
                    type=ObservationSource.KAFKA,
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
        assert cfg.environment.inference_schema == dataclasses.asdict(
            expected_cfg.environment.inference_schema
        )
        assert cfg.environment.observability_backend == dataclasses.asdict(
            expected_cfg.environment.observability_backend
        )
        assert cfg.environment.observation_source == dataclasses.asdict(
            expected_cfg.environment.observation_source
        )
