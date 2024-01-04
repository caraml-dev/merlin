from datetime import datetime

import pandas as pd
from arize.pandas.logger import Client
from arize.utils.types import (
    Environments,
    ModelTypes as ArizeModelType,
    Schema as ArizeSchema,
)
from flytekit import current_context, task, Secret
from merlin.observability.inference import (
    InferenceSchema,
    ObservationType,
    BinaryClassificationOutput,
    RankingOutput,
    RegressionOutput,
    PredictionOutput,
)


def get_arize_model_type(prediction_output: PredictionOutput) -> ArizeModelType:
    if isinstance(prediction_output, BinaryClassificationOutput):
        return ArizeModelType.BINARY_CLASSIFICATION
    elif isinstance(prediction_output, RegressionOutput):
        return ArizeModelType.REGRESSION
    elif isinstance(prediction_output, RankingOutput):
        return ArizeModelType.RANKING
    else:
        raise ValueError(f"Unknown prediction output type: {type(prediction_output)}")


def get_prediction_attributes(prediction_output: PredictionOutput) -> dict:
    if isinstance(prediction_output, BinaryClassificationOutput):
        return dict(
            prediction_label_column_name=prediction_output.prediction_label_column,
            prediction_score_column_name=prediction_output.prediction_score_column,
        )
    elif isinstance(prediction_output, RegressionOutput):
        return dict(
            prediction_score_column_name=prediction_output.prediction_score_column,
        )
    elif isinstance(prediction_output, RankingOutput):
        return dict(
            rank_column_name=prediction_output.rank_column,
        )
    else:
        raise ValueError(f"Unknown prediction output type: {type(prediction_output)}")


def get_ground_truth_attributes(prediction_output: PredictionOutput) -> dict:
    if isinstance(prediction_output, BinaryClassificationOutput):
        return dict(
            actual_label_column_name=prediction_output.actual_label_column,
        )
    elif isinstance(prediction_output, RegressionOutput):
        return dict(
            actual_score_column_name=prediction_output.actual_score_column,
        )
    elif isinstance(prediction_output, RankingOutput):
        return dict(
            rank_column=prediction_output.rank_column,
        )
    else:
        raise ValueError(f"Unknown prediction output type: {type(prediction_output)}")


def get_arize_training_schema(
    inference_schema: InferenceSchema, include_prediction_and_ground_truth: bool
) -> ArizeSchema:
    schema_attributes = dict(
        tag_column_names=inference_schema.tag_columns,
    )
    if include_prediction_and_ground_truth:
        schema_attributes |= dict(
            prediction_id_column_name=inference_schema.prediction_id_column,
        )
    schema_attributes |= dict(
        feature_column_names=inference_schema.feature_columns,
    )
    schema_attributes |= get_prediction_attributes(
        inference_schema.model_prediction_output
    )
    schema_attributes |= get_ground_truth_attributes(
        inference_schema.model_prediction_output
    )
    return ArizeSchema(**schema_attributes)


def add_default_prediction_and_ground_truth_columns(
    inference_schema: InferenceSchema, df: pd.DataFrame
):
    prediction_output = inference_schema.model_prediction_output
    if isinstance(prediction_output, BinaryClassificationOutput):
        df[prediction_output.prediction_score_column] = 0.0
        df[
            prediction_output.actual_label_column
        ] = prediction_output.negative_class_label
    elif isinstance(prediction_output, RegressionOutput):
        df[prediction_output.prediction_score_column] = 0.0
        df[prediction_output.actual_score_column] = 0.0
    elif isinstance(prediction_output, RankingOutput):
        df[prediction_output.rank_column] = (
            df.groupby(prediction_output.prediction_group_id_column).cumcount() + 1
        )
        df[prediction_output.relevance_score_column] = 0.0
    else:
        raise ValueError(f"Unknown prediction output type: {type(prediction_output)}")


ARIZE_SECRET_GROUP = "mlobs"
ARIZE_SPACE_KEY_SECRET_NAME = "arize_space_key"
ARIZE_API_KEY_SECRET_NAME = "arize_api_key"


@task(
    secret_requests=[
        Secret(ARIZE_SECRET_GROUP, ARIZE_SPACE_KEY_SECRET_NAME),
        Secret(ARIZE_SECRET_GROUP, ARIZE_API_KEY_SECRET_NAME),
    ]
)
def publish_training_data_to_arize(
    df: pd.DataFrame,
    inference_schema: dict,
    model_id: str,
    model_version: str,
    include_prediction_and_ground_truth: bool,
):
    inference_schema = InferenceSchema.from_dict(inference_schema)
    space_key = current_context().secrets.get(
        ARIZE_SECRET_GROUP, ARIZE_SPACE_KEY_SECRET_NAME
    )
    api_key = current_context().secrets.get(
        ARIZE_SECRET_GROUP, ARIZE_API_KEY_SECRET_NAME
    )
    client = Client(space_key=space_key, api_key=api_key)
    if not include_prediction_and_ground_truth:
        add_default_prediction_and_ground_truth_columns(inference_schema, df)
    schema = get_arize_training_schema(
        inference_schema, include_prediction_and_ground_truth
    )
    arize_environment = Environments.TRAINING
    processed_df = inference_schema.model_prediction_output.preprocess(
        df,
        [
            ObservationType.FEATURE,
            ObservationType.PREDICTION,
            ObservationType.GROUND_TRUTH,
        ],
    )
    processed_df.reset_index(drop=True, inplace=True)
    version = f"{model_version}-{datetime.now().strftime('%Y%m%d')}-{current_context().execution_id.name}"
    arize_model_types = get_arize_model_type(inference_schema.model_prediction_output)

    client.log(
        dataframe=processed_df,
        environment=arize_environment,
        schema=schema,
        model_id=model_id,
        model_type=arize_model_types,
        model_version=version,
    )
