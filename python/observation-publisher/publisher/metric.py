from pandas import Timestamp
from prometheus_client import Counter, Gauge


class MetricWriter(object):
    """
    Singleton class for writing metrics to Prometheus.
    """

    _instance = None

    def __init__(self):
        if not self._initialized:
            self.model_id = None
            self.model_version = ""
            self.last_processed_timestamp_gauge = Gauge(
                "last_processed_timestamp",
                "The timestamp of the last prediction log processed by the publisher",
                ["model_id", "model_version"],
            )
            self.total_prediction_logs_processed_counter = Counter(
                "total_prediction_logs_processed",
                "The total number of prediction logs processed by the publisher",
                ["model_id", "model_version"],
            )
            self.total_error_process_prediction_logs_counter = Counter(
                "total_error_process_prediction_logs",
                "The total number of prediction logs encounter error during procesing",
                ["model_id", "model_version"],
            )
            self.kafka_consumer_lag_gauge = Gauge(
                "kafka_consumer_lag",
                "The number of unprocess message in kafka",
                ["model_id", "model_version", "partition"],
            )
            self._initialized = True

    def __new__(cls):
        if not cls._instance:
            cls._instance = super(MetricWriter, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def setup(self, model_id: str, model_version: str):
        """
        Needs to be run before sending metrics, so that the singleton instance has the correct properties value.
        :param model_id:
        :param model_version:
        :return:
        """
        self.model_id = model_id
        self.model_version = model_version

    def update_last_processed_timestamp(self, last_processed_timestamp: Timestamp):
        """
        Updates the last_processed_timestamp gauge with the given value.
        :param last_processed_timestamp:
        :return:
        """
        self.last_processed_timestamp_gauge.labels(
            model_id=self.model_id, model_version=self.model_version
        ).set(last_processed_timestamp.timestamp())

    def increment_total_prediction_logs_processed(self, value: int):
        """
        Increments the total_prediction_logs_processed counter by value.
        :return:
        """
        self.total_prediction_logs_processed_counter.labels(
            model_id=self.model_id, model_version=self.model_version
        ).inc(value)

    def increment_total_error_process_prediction_logs(self):
        """
        Increments the total_error_process_prediction_logs counter by value.
        :return:
        """
        self.total_error_process_prediction_logs_counter.labels(
            model_id=self.model_id, model_version=self.model_version
        ).inc()

    def update_kafka_lag(self, total_lag: int, partition: int):
        """
        Update the kafka_consumer_lag gauge with the given value
        :param total_lag:
        :param partition:
        :return:
        """
        self.kafka_consumer_lag_gauge.labels(
            model_id=self.model_id,
            model_version=self.model_version,
            partition=partition,
        ).set(total_lag)
