import uuid

from pyfuncserver.config import Publisher as PublisherConfig, ModelManifest
from confluent_kafka import Producer
from merlin.pyfunc import PyFuncOutput
from caraml.upi.v1 import prediction_log_pb2


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

    def produce(self, data: PyFuncOutput):
        model_input = data.to_model_input_proto()
        model_output = data.to_model_output_proto()
        

        session_id = data.get_session_id()
        prediction_log = prediction_log_pb2.PredictionLog(
            prediction_id=session_id,
            target_name="", # TO-DO update this after schema is introduced
            project_name=self.model_manifest.project,
            model_name=self.model_manifest.model_name,
            model_version=self.model_manifest.model_version,
            input=model_input,
            output=model_output,
            table_schema_version=1
        )
        serialized_data = prediction_log.SerializeToString() 
        self.producer.produce(topic=self.topic, value=serialized_data)
