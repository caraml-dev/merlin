import hydra
from merlin.observability.inference import InferenceSchema
from omegaconf import OmegaConf

from publisher.config import PublisherConfig
from publisher.observability_backend import new_observation_sink
from publisher.prediction_log_consumer import new_consumer


@hydra.main(version_base=None, config_path="../conf", config_name="config")
def start_consumer(cfg: PublisherConfig) -> None:
    missing_keys: set[str] = OmegaConf.missing_keys(cfg)
    if missing_keys:
        raise RuntimeError(f"Got missing keys in config:\n{missing_keys}")
    prediction_log_consumer = new_consumer(cfg.environment.observation_source)
    inference_schema = InferenceSchema.from_dict(
        OmegaConf.to_container(cfg.environment.inference_schema)
    )
    observation_sink = new_observation_sink(
        config=cfg.environment.observability_backend,
        inference_schema=inference_schema,
        model_id=cfg.environment.model_id,
        model_version=cfg.environment.model_version,
    )
    prediction_log_consumer.start_polling(
        observation_sink=observation_sink,
        inference_schema=inference_schema,
    )


if __name__ == "__main__":
    start_consumer()
