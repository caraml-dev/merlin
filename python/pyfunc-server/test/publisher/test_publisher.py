import pytest
from unittest.mock import MagicMock
from pyfuncserver.publisher.publisher import Producer, Publisher
from pyfuncserver.sampler.sampler import Sampler
from test.test_http import sample_model_input, sample_model_output
from merlin.pyfunc import PyFuncOutput

class AlwaysSampling(Sampler):
    def should_sample(self, request_id: int = None) -> bool :
        return True
    
class NeverSampling(Sampler):
    def should_sample(self, request_id: int = None) -> bool :
        return False


@pytest.mark.asyncio
@pytest.mark.parametrize("sampler, is_publishing", [(AlwaysSampling(), True), (NeverSampling(), False)])
async def test_publisher(sampler, is_publishing):
    producer = MagicMock(spec=Producer)
    publisher = Publisher(producer, sampler)
    pyfunc_output = PyFuncOutput(
        http_response = {"response": "ok"},
        model_input=sample_model_input,
        model_output=sample_model_output
    )
    await publisher.publish(pyfunc_output)
    if is_publishing:
        producer.produce.assert_called_once()
    else:
        producer.produce.assert_not_called()
