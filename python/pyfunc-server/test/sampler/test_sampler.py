import pytest
import math

from pyfuncserver.sampler.sampler import RatioSampling

@pytest.mark.parametrize("num_requests, ratio", [(3000, 0.2), (10000, 0.3), (5000, 0.5)])
def test_ratio_sampling(num_requests, ratio):
    ratio_based_sampler = RatioSampling(ratio=ratio)
    num_sampled = 0
    for _ in range(num_requests):
        if ratio_based_sampler.should_sample():
            num_sampled = num_sampled + 1
    actual_ratio = num_sampled / num_requests
    print(f"actual ratio {actual_ratio} expected ratio {ratio}")
    assert math.isclose(a=ratio, b=actual_ratio, abs_tol=0.02)
