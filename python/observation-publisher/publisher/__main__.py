import hydra
from omegaconf import OmegaConf

from publisher.config import PublisherConfig
from publisher.observability_backend import new_observation_sink
from publisher.observation_log_consumer import new_consumer


@hydra.main(version_base=None, config_path="../conf", config_name="config")
def start_consumer(cfg: PublisherConfig) -> None:
    missing_keys: set[str] = OmegaConf.missing_keys(cfg)
    if missing_keys:
        raise RuntimeError(f"Got missing keys in config:\n{missing_keys}")
    prediction_log_consumer = new_consumer(cfg.environment.observation_source)
    observation_sink = new_observation_sink(
        cfg.environment.observability_backend, cfg.environment.model
    )
    prediction_log_consumer.start_polling(
        observation_sink=observation_sink, model_spec=cfg.environment.model
    )


if __name__ == "__main__":
    start_consumer()
