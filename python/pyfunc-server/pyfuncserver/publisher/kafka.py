from pyfuncserver.config import Publisher as PublisherConfig, ModelManifest
from pyfuncserver.utils.converter import build_prediction_log

from confluent_kafka import KafkaError, KafkaException, Producer
from confluent_kafka.admin import AdminClient, NewTopic
from merlin.pyfunc import PyFuncOutput


class KafkaProducer(Producer):
    def __init__(self, publisher_config: PublisherConfig, model_manifest: ModelManifest) -> None:
        conf = {
            "bootstrap.servers": publisher_config.kafka.brokers, 
            "acks": publisher_config.kafka.acks, 
            "linger.ms": publisher_config.kafka.linger_ms
        }
        conf.update(publisher_config.kafka.configuration)
        self.producer = Producer(**conf)
        self.topic = publisher_config.kafka.topic
        self.model_manifest = model_manifest
        admin_client = AdminClient(conf)
        try:
            admin_client.create_topics([
                NewTopic(
                    self.topic,
                    num_partitions=publisher_config.num_partitions,
                    replication_factor=publisher_config.replication_factor
                )
            ])[self.topic].result()
        except KafkaException as e:
            if e.args[0].code() != KafkaError.TOPIC_ALREADY_EXISTS:
                raise e

    def produce(self, data: PyFuncOutput):
        prediction_log = build_prediction_log(pyfunc_output=data, model_manifest=self.model_manifest)
        serialized_data = prediction_log.SerializeToString() 
        self.producer.produce(topic=self.topic, value=serialized_data)
        self.producer.poll(0)
