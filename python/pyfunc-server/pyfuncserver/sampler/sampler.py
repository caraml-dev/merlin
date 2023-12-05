import random
from abc import ABC, abstractmethod
from typing import Optional

class Sampler(ABC):
    @abstractmethod
    def should_sample(self, request_id: Optional[int] = None) -> bool:
        pass

class RatioSampling(Sampler):
    """
    RatioSampling sampling strategy that mimic implementation of the TraceIDRatioBased in opentelemetry
    ref: https://github.com/open-telemetry/opentelemetry-python/blob/v1.21.0/opentelemetry-sdk/src/opentelemetry/sdk/trace/sampling.py#L253
    """
    REQUEST_ID_LIMIT = (1 << 64) - 1

    def __init__(self, ratio: float) -> None:
        if ratio < 0.0 or ratio > 1.0:
            raise ValueError("Valid ratio is on 0 - 1 range")
        self._ratio = ratio   
        self._bound = self.get_bound_for_ratio(ratio)

    @classmethod
    def get_bound_for_ratio(cls, ratio: float) -> int:
        return round(ratio * (cls.REQUEST_ID_LIMIT + 1))
    
    @property
    def ratio(self) -> float:
        return self._ratio
    
    def should_sample(self, request_id: Optional[int] = None) -> bool :
        if request_id is None:
            request_id = random.getrandbits(128)
        
        return request_id & self.REQUEST_ID_LIMIT < self._bound
