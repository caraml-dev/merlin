import random

class RatioSampling:
    
    def __init__(self, ratio: float) -> None:
        if ratio < 0.0 or ratio > 1.0:
            raise ValueError("Valid ratio is on 0 - 1 range")
        self._ratio = ratio   
        self._bound = self.get_bound_for_ratio(ratio)

    REQUEST_ID_LIMIT = (1 << 64) - 1

    @classmethod
    def get_bound_for_ratio(cls, ratio: float) -> int:
        return round(ratio * (cls.REQUEST_ID_LIMIT + 1))
    
    @property
    def ratio(self) -> float:
        return self._ratio
    
    def should_sample(self, request_id: int = None):
        if request_id is None:
            request_id = random.getrandbits(128)
        
        return request_id & self.REQUEST_ID_LIMIT < self._bound
