import abc
from typing import Any

import farmhash
import pandas as pd
from flytekit import task
from merlin.observability.inference import InferenceSchema


class Sampler(abc.ABC):
    @abc.abstractmethod
    def sample(self, df: pd.DataFrame, sampling_rate: float) -> pd.DataFrame:
        raise NotImplementedError


class RandomSampler(Sampler):
    def sample(self, df: pd.DataFrame, sampling_rate: float) -> pd.DataFrame:
        return df.sample(frac=sampling_rate)


def generate_farmhash_fingerprint(value: Any) -> int:
    return farmhash.fingerprint64(str(value))


class ConsistentSampler(Sampler):
    def __init__(self, hash_column: str):
        self._hash_column = hash_column

    def sample(self, df: pd.DataFrame, sampling_rate: float) -> pd.DataFrame:
        filtered = df[self._hash_column].apply(
            lambda x: (generate_farmhash_fingerprint(x) % 100)
            <= int(100 * sampling_rate),
            axis=1,
        )
        return df[filtered]


@task
def sample_dataframe(
    df: pd.DataFrame,
    sampling_rate: float,
    sampling_strategy: str,
    inference_schema: dict,
) -> pd.DataFrame:
    inference_schema = InferenceSchema.from_dict(inference_schema)
    match sampling_strategy:
        case "random":
            sampler = RandomSampler()
        case "consistent":
            sampler = ConsistentSampler(
                hash_column=inference_schema.prediction_id_column
            )
        case _:
            raise ValueError(f"Invalid sampling strategy: {sampling_strategy}")
    return sampler.sample(df=df, sampling_rate=sampling_rate)
