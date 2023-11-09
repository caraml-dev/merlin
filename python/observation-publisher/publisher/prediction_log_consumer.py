from typing import List, Optional, Union

import numpy as np
import pandas as pd
from caraml.upi.v1.prediction_log_pb2 import PredictionLog
from confluent_kafka import Consumer, KafkaException
from google.protobuf.internal.well_known_types import ListValue

from publisher.config import (KafkaConsumerConfig, ModelSchema, ModelSpec,
                              ValueType)
from publisher.observability_backend import ObservationSink


def parse_message_to_prediction_log(msg: str) -> PredictionLog:
    log = PredictionLog()
    log.ParseFromString(msg)
    return log


def poll_prediction_logs(
    kafka_consumer: Consumer,
    consumer_cfg: KafkaConsumerConfig,
    observation_sink: ObservationSink,
    model_cfg: ModelSpec,
):
    kafka_consumer.subscribe([consumer_cfg.topic])
    while True:
        messages = kafka_consumer.consume(
            consumer_cfg.batch_size, timeout=consumer_cfg.poll_timeout_seconds
        )
        errors = [msg.error() for msg in messages if msg.error() is not None]

        if len(errors) > 0:
            print(f"Last encountered error: {errors[-1]}")
            raise KafkaException(errors[-1])
        prediction_logs = [
            parse_message_to_prediction_log(msg.value().decode("utf-8"))
            for msg in messages
            if (msg is not None and msg.error() is None)
        ]
        if len(prediction_logs) > 0:
            df = log_to_dataframe(prediction_logs, model_cfg.schema)
            observation_sink.write(dataframe=df)
            kafka_consumer.commit()


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


def convert_list_value(list_value: ListValue, value_type: List[ValueType]) -> List:
    return [
        convert_value(col_value, col_type)
        for col_value, col_type in zip([v for v in list_value], value_type)
    ]


def log_to_dataframe(
    logs: List[PredictionLog], model_schema: ModelSchema
) -> pd.DataFrame:
    combined_records = []
    column_names = []
    for log in logs:
        request_timestamp = log.request_timestamp.ToDatetime()
        rows = [
            convert_list_value(feature_row, model_schema.feature_column_types)
            + convert_list_value(prediction_row, model_schema.prediction_column_types)
            + [log.prediction_id + row_id, request_timestamp]
            for feature_row, prediction_row, row_id in zip(
                log.input.features_table["data"],
                log.output.prediction_results_table["data"],
                log.output.prediction_results_table["row_ids"],
            )
        ]
        column_names = (
            [c for c in log.input.features_table["column_names"]]
            + [c for c in log.output.prediction_results_table["column_names"]]
            + ["prediction_id", "request_timestamp"]
        )
        combined_records.extend(rows)
    return pd.DataFrame.from_records(combined_records, columns=column_names)
