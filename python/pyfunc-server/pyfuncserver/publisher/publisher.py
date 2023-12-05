from pyfuncserver.sampler.sampler import Sampler
from merlin.pyfunc import PyFuncOutput
from abc import ABC, abstractmethod
import asyncio

class Producer(ABC):

    @abstractmethod
    def produce(self, data: PyFuncOutput):
        pass

class Publisher:
    def __init__(self, producer: Producer, sampler: Sampler) -> None:
        self.producer = producer
        self.sampler = sampler

    async def publish(self, output: PyFuncOutput):
        if not self.sampler.should_sample():
            return

        self.producer.produce(output)