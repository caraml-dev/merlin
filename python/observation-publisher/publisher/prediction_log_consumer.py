import abc
from typing import List, Tuple

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from confluent_kafka import Consumer, KafkaException
from merlin.observability.inference import InferenceSchema

from publisher.config import (
    KafkaConsumerConfig,
    ObservationSource,
    ObservationSourceConfig,
)
from publisher.observability_backend import ObservationSink
from publisher.prediction_log_parser import (
    PREDICTION_LOG_TIMESTAMP_COLUMN,
    parse_struct_to_feature_table,
    parse_struct_to_result_table,
)


class PredictionLogConsumer(abc.ABC):
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
        self, observation_sink: ObservationSink, inference_schema: InferenceSchema
    ):
        try:
            while True:
                logs = self.poll_new_logs()
                if len(logs) == 0:
                    continue
                df = log_batch_to_dataframe(logs, inference_schema)
                observation_sink.write(df)
                self.commit()
        finally:
            self.close()


class KafkaPredictionLogConsumer(PredictionLogConsumer):
    def __init__(self, config: KafkaConsumerConfig):
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
        return KafkaPredictionLogConsumer(config.kafka_config)
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
    feature_table = parse_struct_to_feature_table(
        log.input.features_table, inference_schema
    )
    prediction_results_table = parse_struct_to_result_table(
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
    column_names = []
    for log in logs:
        rows, column_names = log_to_records(log, inference_schema)
        combined_records.extend(rows)
    return pd.DataFrame.from_records(combined_records, columns=column_names)
