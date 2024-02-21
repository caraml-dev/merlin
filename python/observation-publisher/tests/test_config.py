import dataclasses

from hydra import compose, initialize
from merlin.observability.inference import InferenceSchema, ValueType
from omegaconf import OmegaConf

from publisher.config import (Environment, ObservationSinkConfig,
                              ObservationSinkType, ObservationSource,
                              ObservationSourceConfig, PublisherConfig)


def test_config_initialization():
    with initialize(version_base=None, config_path="../conf"):
        cfg = compose(config_name="config", overrides=["+environment=example-override"])
        expected_cfg = PublisherConfig(
            environment=Environment(
                model_id="test-model",
                model_version="0.1.0",
                inference_schema=dict(
                    feature_types={
                        "distance": ValueType.INT64,
                        "transaction": ValueType.FLOAT64,
                    },
                    model_prediction_output=dict(
                        output_class="BinaryClassificationOutput",
                        prediction_score_column="score",
                        actual_score_column="actual_score",
                        positive_class_label="positive",
                        negative_class_label="negative",
                        score_threshold=0.5,
                    ),
                ),
                observation_sinks=[
                    ObservationSinkConfig(
                        type=ObservationSinkType.ARIZE,
                        config=dict(
                            api_key="SECRET_API_KEY",
                            space_key="SECRET_SPACE_KEY",
                        ),
                    )
                ],
                observation_source=ObservationSourceConfig(
                    type=ObservationSource.KAFKA,
                    config=dict(
                        topic="test-topic",
                        bootstrap_servers="localhost:9092",
                        group_id="test-group",
                        batch_size=100,
                        additional_consumer_config={
                            "auto.offset.reset": "latest",
                        },
                    ),
                ),
            )
        )
        parsed_schema: InferenceSchema = InferenceSchema.from_dict(
            OmegaConf.to_container(cfg.environment.inference_schema)
        )
        assert parsed_schema == InferenceSchema.from_dict(
            expected_cfg.environment.inference_schema
        )
        assert cfg.environment.observation_sinks[0] == dataclasses.asdict(
            expected_cfg.environment.observation_sinks[0]
        )
        assert cfg.environment.observation_source == dataclasses.asdict(
            expected_cfg.environment.observation_source
        )
