import pandas as pd
from flytekit import dynamic, Resources

from publisher.ingestion import publish_training_data_to_arize
from publisher.sampling import sample_dataframe


@dynamic
def ingest_training_data(
    df: pd.DataFrame,
    model_id: str,
    model_version: str,
    inference_schema: dict,
    include_prediction_and_ground_truth: bool,
    sampling_rate: float,
    sampling_strategy: str,
    cpu: str,
    memory: str,
):
    """
    Ingests batch data to Arize.

    :param df: batch data.
    :param inference_schema: Inference schema for the given model id and version
    :param model_id: Model id.
    :param model_version: Model version.
    :param include_prediction_and_ground_truth: Send prediction and ground truth together with the features.
        If false, only send the features.
    :param sampling_rate: Sampling rate (0.0 - 1.0).
    :param sampling_strategy: One of ("random", "consistent"). Random sampling will randomly sample the data.
        Consistent sampling is deterministic, but will round the sampling rate to the closest percentage integer.
    :param memory: Memory request for the ingestion task
    :param cpu: CPU request for the ingestion task
    :return:
    """
    sampled_dataframe = df
    if sampling_rate > 0.0:
        sampled_dataframe = sample_dataframe(
            df=df,
            sampling_rate=sampling_rate,
            sampling_strategy=sampling_strategy,
            inference_schema=inference_schema,
        ).with_overrides(
            requests=Resources(cpu=cpu, mem=memory),
            limits=Resources(cpu=cpu, mem=memory),
        )

    publish_training_data_to_arize(
        df=sampled_dataframe,
        inference_schema=inference_schema,
        model_id=model_id,
        model_version=model_version,
        include_prediction_and_ground_truth=include_prediction_and_ground_truth,
    ).with_overrides(
        requests=Resources(cpu=cpu, mem=memory), limits=Resources(cpu=cpu, mem=memory)
    )
