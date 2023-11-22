import pytest
from unittest.mock import MagicMock, patch, ANY
from pyfuncserver.publisher.kafka import KafkaProducer
from pyfuncserver.config import ModelManifest

from test.test_http import sample_model_input, sample_model_output
from merlin.pyfunc import PyFuncOutput
from test.test_http import sample_model_input, sample_model_output

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
        producer.produce.assert_called_once_with(topic="sample_topic", value=ANY)
        