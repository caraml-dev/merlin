import hydra
from merlin.observability.inference import InferenceSchema
from omegaconf import OmegaConf
from prometheus_client import start_http_server

from publisher.config import PublisherConfig
from publisher.metric import MetricWriter
from publisher.observation_sink import new_observation_sink
from publisher.prediction_log_consumer import new_consumer


@hydra.main(version_base=None, config_path="../conf", config_name="config")
def start_consumer(cfg: PublisherConfig) -> None:
    missing_keys: set[str] = OmegaConf.missing_keys(cfg)
    if missing_keys:
        raise RuntimeError(f"Got missing keys in config:\n{missing_keys}")

    start_http_server(cfg.environment.prometheus_port)
    MetricWriter().setup(
        model_id=cfg.environment.model_id, model_version=cfg.environment.model_version
    )
    prediction_log_consumer = new_consumer(cfg.environment.observation_source)
    inference_schema = InferenceSchema.from_dict(
        OmegaConf.to_container(cfg.environment.inference_schema)
    )
    observation_sinks = [
        new_observation_sink(
            sink_config=sink_config,
            inference_schema=inference_schema,
            model_id=cfg.environment.model_id,
            model_version=cfg.environment.model_version,
        )
        for sink_config in cfg.environment.observation_sinks
    ]
    prediction_log_consumer.start_polling(
        observation_sinks=observation_sinks,
        inference_schema=inference_schema,
    )


if __name__ == "__main__":
    start_consumer()
