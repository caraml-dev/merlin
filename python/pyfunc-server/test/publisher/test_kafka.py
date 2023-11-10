import pytest
from unittest.mock import MagicMock, patch
from pyfuncserver.publisher.kafka import KafkaProducer
from pyfuncserver.config import ModelManifest

from test.test_http import sample_model_input, sample_model_output
from merlin.pyfunc import PyFuncOutput
from test.test_http import sample_model_input, sample_model_output
from caraml.upi.v1 import prediction_log_pb2, upi_pb2, table_pb2, type_pb2

@pytest.mark.parametrize("data", [
    PyFuncOutput(
        http_response = {"response": "ok"},
        model_input=sample_model_input,
        model_output=sample_model_output
    )
])
def test_kafka_producer(data):
    with patch.object(KafkaProducer, "__init__", lambda x, y, z: None):
        kafka_producer = KafkaProducer(None, None)
        producer = MagicMock()
        kafka_producer.producer = producer
        kafka_producer.topic = "sample_topic"
        kafka_producer.model_manifest = ModelManifest(model_name="model_name", model_version="1", model_full_name="model_name_1", model_dir="/", project="project")
        
        kafka_producer.produce(data)
        exp_prediction_log = prediction_log_pb2.PredictionLog(
            prediction_id=sample_model_input.session_id,
            target_name="",
            project_name="project",
            model_name="model_name",
            model_version="1",
            input = data.to_model_input_proto(),
            output = data.to_model_output_proto(),
            table_schema_version=1
        )
        serialized_data = exp_prediction_log.SerializeToString() 
        producer.produce.assert_called_once_with(topic="sample_topic", value=serialized_data)
        