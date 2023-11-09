import hydra
from confluent_kafka import Consumer
from omegaconf import OmegaConf

from publisher.config import PublisherConfig
from publisher.observability_backend import new_observation_sink
from publisher.prediction_log_consumer import poll_prediction_logs


@hydra.main(version_base=None, config_path="../conf", config_name="config")
def start_consumer(cfg: PublisherConfig) -> None:
    missing_keys: set[str] = OmegaConf.missing_keys(cfg)
    if missing_keys:
        raise RuntimeError(f"Got missing keys in config:\n{missing_keys}")
    kafka_consumer_cfg = cfg.environment.kafka
    kafka_consumer = Consumer(
        {
            "bootstrap.servers": kafka_consumer_cfg.bootstrap_servers,
            "group.id": kafka_consumer_cfg.group_id,
            "auto.offset.reset": kafka_consumer_cfg.auto_offset_reset,
            "enable.partition.eof": False,
        }
    )
    model_cfg = cfg.environment.model
    observation_sink = new_observation_sink(
        cfg.environment.observability_backend, model_cfg
    )

    try:
        poll_prediction_logs(
            kafka_consumer=kafka_consumer,
            observation_sink=observation_sink,
            consumer_cfg=kafka_consumer_cfg,
            model_cfg=model_cfg,
        )
    finally:
        kafka_consumer.close()


if __name__ == "__main__":
    start_consumer()
