import pytest
import pandas as pd

from merlin.observability.inference import BinaryClassificationOutput, PredictionOutput, ObservationType, RankingOutput


@pytest.fixture
def binary_classification_output() -> BinaryClassificationOutput:
    return BinaryClassificationOutput(
        prediction_score_column="prediction_score",
        actual_label_column="target",
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
        "actual_label_column": "target",
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
            [0.8, "ACCEPTED"],
            [0.5, "ACCEPTED"],
            [1.0, "REJECTED"],
            [0.4, "ACCEPTED"]
        ],
        columns=[
            "prediction_score",
            "target"
        ],
    )
    processed_df = binary_classification_output.preprocess(input_df, [ObservationType.PREDICTION, ObservationType.GROUND_TRUTH])
    pd.testing.assert_frame_equal(
        processed_df,
        pd.DataFrame.from_records(
            [
                [0.8, "ACCEPTED" , "ACCEPTED", 1.0],
                [0.5, "ACCEPTED", "ACCEPTED", 1.0],
                [1.0, "REJECTED", "ACCEPTED", 0.0],
                [0.4, "ACCEPTED", "REJECTED", 1.0]
            ],
            columns=[
                "prediction_score",
                "target",
                "_prediction_label",
                "_actual_score",
            ],
        ),
    )


@pytest.fixture
def ranking_prediction_output() -> RankingOutput:
    return RankingOutput(
        rank_score_column="score",
        prediction_group_id_column="order_id",
        relevance_score_column="relevance_score_column",
    )


@pytest.mark.unit
def test_ranking_preprocessing(ranking_prediction_output: RankingOutput):
    input_df = pd.DataFrame.from_records(
        [
            [1.0, "1001"],
            [0.8, "1001"],
            [0.1, "1001"],
            [0.3, "1002"],
            [0.2, "1002"],
            [0.1, "1002"],
        ],
        columns=[
            "score",
            "order_id",
        ],
    )
    processed_df = ranking_prediction_output.preprocess(input_df, [ObservationType.PREDICTION])
    pd.testing.assert_frame_equal(
        processed_df,
        pd.DataFrame.from_records(
            [
                [1.0, "1001", 1],
                [0.8, "1001", 2],
                [0.1, "1001", 3],
                [0.3, "1002", 1],
                [0.2, "1002", 2],
                [0.1, "1002", 3],
            ],
            columns=[
                "score",
                "order_id",
                "_rank",
            ],
        ),
        check_like=True
    )
