import pytest
import pandas as pd

from merlin.observability.inference import BinaryClassificationOutput, PredictionOutput, ObservationType, RankingOutput


@pytest.fixture
def binary_classification_output() -> BinaryClassificationOutput:
    return BinaryClassificationOutput(
        prediction_score_column="prediction_score",
        actual_score_column="actual_score",
        positive_class_label="ACCEPTED",
        negative_class_label="REJECTED",
        score_threshold=0.5,
    )


@pytest.mark.unit
def test_prediction_output_encoding(binary_classification_output: BinaryClassificationOutput):
    encoded_output = PredictionOutput.encode_with_discriminator(binary_classification_output)
    assert encoded_output == {
        "output_class": "BinaryClassificationOutput",
        "prediction_score_column": "prediction_score",
        "actual_score_column": "actual_score",
        "positive_class_label": "ACCEPTED",
        "negative_class_label": "REJECTED",
        "score_threshold": 0.5,
    }
    decoded_output = PredictionOutput.decode(encoded_output)
    assert isinstance(decoded_output, BinaryClassificationOutput)
    assert decoded_output == binary_classification_output


@pytest.mark.unit
def test_binary_classification_preprocessing(binary_classification_output: BinaryClassificationOutput):
    input_df = pd.DataFrame.from_records(
        [
            ["9001", "1001", 0.8, 1.0],
            ["9002", "1001", 0.5, 1.0],
            ["9002", "1002", 1.0, 0.0],
            ["9003", "1003", 0.4, 1.0]
        ],
        columns=[
            "order_id",
            "driver_id",
            "prediction_score",
            "actual_score"
        ],
    )
    processed_df = binary_classification_output.preprocess(input_df, "order_id", "driver_id", [ObservationType.PREDICTION, ObservationType.GROUND_TRUTH])
    pd.testing.assert_frame_equal(
        processed_df,
        pd.DataFrame.from_records(
            [
                ["9001", "1001", 0.8, 1.0, "ACCEPTED" , "ACCEPTED"],
                ["9002", "1001", 0.5, 1.0, "ACCEPTED", "ACCEPTED"],
                ["9002", "1002", 1.0, 0.0, "ACCEPTED", "REJECTED"],
                ["9003", "1003", 0.4, 1.0, "REJECTED", "ACCEPTED"]
            ],
            columns=[
                "order_id",
                "driver_id",
                "prediction_score",
                "actual_score",
                "_prediction_label",
                "_actual_label",
            ],
        ),
    )


@pytest.fixture
def ranking_prediction_output() -> RankingOutput:
    return RankingOutput(
        rank_score_column="score",
        relevance_score_column="relevance_score_column",
    )


@pytest.mark.unit
def test_ranking_preprocessing(ranking_prediction_output: RankingOutput):
    input_df = pd.DataFrame.from_records(
        [
            [1.0, "1001", "d1"],
            [0.8, "1001", "d2"],
            [0.1, "1001", "d3"],
            [0.3, "1002", "d1"],
            [0.2, "1002", "d2"],
            [0.1, "1002", "d3"],
        ],
        columns=[
            "score",
            "order_id",
            "driver_id"
        ],
    )
    processed_df = ranking_prediction_output.preprocess(input_df, "order_id", "driver_id", [ObservationType.PREDICTION])
    pd.testing.assert_frame_equal(
        processed_df,
        pd.DataFrame.from_records(
            [
                [1.0, "1001", "d1", 1],
                [0.8, "1001", "d2", 2],
                [0.1, "1001", "d3", 3],
                [0.3, "1002", "d1", 1],
                [0.2, "1002", "d2", 2],
                [0.1, "1002", "d3", 3],
            ],
            columns=[
                "score",
                "order_id",
                "driver_id",
                "_rank",
            ],
        ),
        check_like=True
    )
