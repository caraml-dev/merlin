from pyfuncserver.config import Publisher as PublisherConfig, ModelManifest
from pyfuncserver.publisher.sampler import RatioSampling
from merlin.pyfunc import PyFuncOutput
from caraml.upi.v1 import prediction_log_pb2
import uuid

from confluent_kafka import Producer

class Publisher:

    def __init__(self, publisher_config: PublisherConfig) -> None:
        conf = {
            "bootstrap.servers": publisher_config.kafka.brokers, 
            "acks": publisher_config.kafka.acks, 
            "linger.ms": 1000
        }
        conf.update(publisher_config.kafka.configuration)
        self.producer = Producer(**conf)
        self.sampler = RatioSampling(publisher_config.sampling_ratio)
        self.topic = publisher_config.kafka.topic

    async def publish(self, output: PyFuncOutput, model_manifest: ModelManifest):
        if not self.sampler.should_sample():
            return
        
        model_input = output.to_model_input_proto()
        model_output = output.to_model_output_proto()

        prediction_log = prediction_log_pb2.PredictionLog(
            prediction_id=str(uuid.uuid4()),
            target_name="",
            project_name=model_manifest.project,
            model_name=model_manifest.model_name,
            model_version=model_manifest.model_version,
            input=model_input,
            output=model_output,
            table_schema_version=1
        )
        serialized_data = prediction_log.SerializeToString() 
        self.producer.produce(topic=self.topic, value=serialized_data)



        
        