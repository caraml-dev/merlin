import abc
from dataclasses import dataclass
from datetime import datetime
from threading import Thread
from typing import List, Optional, Tuple

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from confluent_kafka import Consumer, KafkaException
from dataclasses_json import DataClassJsonMixin, dataclass_json
from merlin.observability.inference import InferenceSchema

from publisher.config import ObservationSource, ObservationSourceConfig
from publisher.metric import MetricWriter
from publisher.observation_sink import ObservationSink
from publisher.prediction_log_parser import (
    PREDICTION_LOG_TIMESTAMP_COLUMN,
    PredictionLogFeatureTable,
    PredictionLogResultsTable,
)


class PredictionLogConsumer(abc.ABC):
    def __init__(self, buffer_capacity: int, buffer_max_duration_seconds: int):
        self.buffer_capacity = buffer_capacity
        self.buffer_max_duration_seconds = buffer_max_duration_seconds

    @abc.abstractmethod
    def poll_new_logs(self) -> List[PredictionLog]:
        raise NotImplementedError

    @abc.abstractmethod
    def commit(self):
        raise NotImplementedError

    @abc.abstractmethod
    def close(self):
        raise NotImplementedError

    def start_polling(
        self,
        observation_sinks: List[ObservationSink],
        inference_schema: InferenceSchema,
    ):
        try:
            buffered_logs = []
            buffered_max_duration_seconds = 60
            buffer_start_time = datetime.now()
            while True:
                logs = self.poll_new_logs()
                if len(logs) == 0:
                    continue
                buffered_logs.extend(logs)
                buffered_duration = (datetime.now() - buffer_start_time).seconds
                if (
                    len(buffered_logs) < self.buffer_capacity
                    and buffered_duration < buffered_max_duration_seconds
                ):
                    continue
                df = log_batch_to_dataframe(buffered_logs, inference_schema)
                most_recent_prediction_timestamp = df[
                    PREDICTION_LOG_TIMESTAMP_COLUMN
                ].max()
                MetricWriter().update_last_processed_timestamp(
                    most_recent_prediction_timestamp
                )
                MetricWriter().increment_total_prediction_logs_processed(
                    len(buffered_logs)
                )
                write_tasks = [
                    Thread(target=sink.write, args=(df,)) for sink in observation_sinks
                ]
                for task in write_tasks:
                    task.start()
                for task in write_tasks:
                    task.join()
                self.commit()
                buffered_logs = []
                buffer_start_time = datetime.now()
        finally:
            self.close()


@dataclass_json
@dataclass
class KafkaConsumerConfig:
    topic: str
    bootstrap_servers: str
    group_id: str
    batch_size: int = 100
    poll_timeout_seconds: float = 1.0
    additional_consumer_config: Optional[dict] = None


class KafkaPredictionLogConsumer(PredictionLogConsumer):
    def __init__(
        self,
        buffer_capacity: int,
        buffer_max_duration_seconds: int,
        config: KafkaConsumerConfig,
    ):
        super().__init__(
            buffer_capacity=buffer_capacity,
            buffer_max_duration_seconds=buffer_max_duration_seconds,
        )
        consumer_config = {
            "bootstrap.servers": config.bootstrap_servers,
            "group.id": config.group_id,
            "enable.auto.commit": False,
        }

        if config.additional_consumer_config is not None:
            consumer_config.update(config.additional_consumer_config)

        self._consumer = Consumer(consumer_config)
        self._batch_size = config.batch_size
        self._consumer.subscribe([config.topic])
        self._poll_timeout = config.poll_timeout_seconds

    def poll_new_logs(self) -> List[PredictionLog]:
        messages = self._consumer.consume(self._batch_size, timeout=self._poll_timeout)
        errors = [msg.error() for msg in messages if msg.error() is not None]
        if len(errors) > 0:
            print(f"Last encountered error: {errors[-1]}")
            raise KafkaException(errors[-1])

        return [
            parse_message_to_prediction_log(msg.value())
            for msg in messages
            if (msg is not None and msg.error() is None)
        ]

    def commit(self):
        self._consumer.commit()

    def close(self):
        self._consumer.close()


def new_consumer(config: ObservationSourceConfig) -> PredictionLogConsumer:
    if config.type == ObservationSource.KAFKA:
        assert issubclass(KafkaConsumerConfig, DataClassJsonMixin)
        kafka_consumer_config: KafkaConsumerConfig = KafkaConsumerConfig.from_dict(
            config.config
        )  # type: ignore[attr-defined]
        return KafkaPredictionLogConsumer(
            config.buffer_capacity,
            config.buffer_max_duration_seconds,
            kafka_consumer_config,
        )
    else:
        raise ValueError(f"Unknown consumer type: {config.type}")


def parse_message_to_prediction_log(msg: str) -> PredictionLog:
    log = PredictionLog()
    log.ParseFromString(msg)
    return log


def log_to_records(
    log: PredictionLog, inference_schema: InferenceSchema
) -> Tuple[List[List[np.int64 | np.float64 | np.bool_ | np.str_]], List[str]]:
    request_timestamp = log.request_timestamp.ToDatetime()
    feature_table = PredictionLogFeatureTable.from_struct(
        log.input.features_table, inference_schema
    )
    prediction_results_table = PredictionLogResultsTable.from_struct(
        log.output.prediction_results_table, inference_schema
    )

    rows = [
        feature_row + prediction_row + [log.prediction_id + row_id, request_timestamp]
        for feature_row, prediction_row, row_id in zip(
            feature_table.rows,
            prediction_results_table.rows,
            prediction_results_table.row_ids,
        )
    ]

    column_names = (
        feature_table.columns
        + prediction_results_table.columns
        + [inference_schema.prediction_id_column, PREDICTION_LOG_TIMESTAMP_COLUMN]
    )

    return rows, column_names


def log_batch_to_dataframe(
    logs: List[PredictionLog], inference_schema: InferenceSchema
) -> pd.DataFrame:
    combined_records = []
    column_names: List[str] = []
    for log in logs:
        rows, column_names = log_to_records(log, inference_schema)
        combined_records.extend(rows)
    return pd.DataFrame.from_records(combined_records, columns=column_names)
