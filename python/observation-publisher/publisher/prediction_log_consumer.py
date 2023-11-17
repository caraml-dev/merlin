import abc
from typing import List, Optional, Union, Dict

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from confluent_kafka import Consumer, KafkaException
from google.protobuf.internal.well_known_types import ListValue

from publisher.config import (
    KafkaConsumerConfig,
    ModelSpec,
    ValueType,
    PredictionLogConsumerConfig,
    ConsumerType,
)
from publisher.observability_backend import ObservationSink


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

    def start_polling(self, observation_sink: ObservationSink, model_spec: ModelSpec):
        try:
            while True:
                logs = self.poll_new_logs()
                if len(logs) == 0:
                    continue
                df = log_to_dataframe(logs, model_spec)
                observation_sink.write(dataframe=df)
                self.commit()
        finally:
            self.close()


class KafkaPredictionLogConsumer(PredictionLogConsumer):
    def __init__(self, config: KafkaConsumerConfig):
        consumer_config = {
            "bootstrap.servers": config.bootstrap_servers,
            "group.id": config.group_id,
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
            parse_message_to_prediction_log(msg.value().decode("utf-8"))
            for msg in messages
            if (msg is not None and msg.error() is None)
        ]

    def commit(self):
        self._consumer.commit()

    def close(self):
        self._consumer.close()


def new_consumer(config: PredictionLogConsumerConfig) -> PredictionLogConsumer:
    if config.type == ConsumerType.KAFKA:
        return KafkaPredictionLogConsumer(config.kafka_config)
    else:
        raise ValueError(f"Unknown consumer type: {config.type}")


def parse_message_to_prediction_log(msg: str) -> PredictionLog:
    log = PredictionLog()
    log.ParseFromString(msg)
    return log


def convert_value(
    col_value: Optional[Union[int, str, float, bool]], value_type: ValueType
):
    match value_type:
        case ValueType.INT64:
            return np.int64(col_value)
        case ValueType.FLOAT64:
            return np.float64(col_value)
        case ValueType.BOOLEAN:
            return np.bool_(col_value)
        case ValueType.STRING:
            return np.str_(col_value)
        case _:
            raise ValueError(f"Unknown value type: {value_type}")


def convert_list_value(
    list_value: ListValue, column_names: List[str], column_types: Dict[str, ValueType]
) -> List:
    return [
        convert_value(col_value, column_types[col_name])
        for col_value, col_name in zip([v for v in list_value], column_names)
    ]


def log_to_dataframe(logs: List[PredictionLog], model_spec: ModelSpec) -> pd.DataFrame:
    combined_records = []
    column_names = []
    for log in logs:
        request_timestamp = log.request_timestamp.ToDatetime()
        rows = [
            convert_list_value(
                feature_row,
                log.input.features_table["columns"],
                model_spec.feature_types,
            )
            + convert_list_value(
                prediction_row,
                log.output.prediction_results_table["columns"],
                model_spec.prediction_types,
            )
            + [log.prediction_id + row_id, request_timestamp]
            for feature_row, prediction_row, row_id in zip(
                log.input.features_table["data"],
                log.output.prediction_results_table["data"],
                log.output.prediction_results_table["row_ids"],
            )
        ]
        column_names = (
            [c for c in log.input.features_table["columns"]]
            + [c for c in log.output.prediction_results_table["columns"]]
            + [model_spec.prediction_id_column, model_spec.timestamp_column]
        )
        combined_records.extend(rows)
    return pd.DataFrame.from_records(combined_records, columns=column_names)
