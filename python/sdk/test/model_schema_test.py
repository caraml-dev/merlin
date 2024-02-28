import pytest
import client
from merlin.model_schema import ModelSchema
from merlin.observability.inference import InferenceSchema, ValueType, BinaryClassificationOutput, RegressionOutput, RankingOutput

@pytest.mark.unit
@pytest.mark.parametrize(
    "response,expected,error",
    [
        (
            client.ModelSchema(
                id=1,
                model_id=1,
                spec=client.SchemaSpec(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags"],
                    feature_types={
                        "featureA": client.ValueType.FLOAT64,
                        "featureB": client.ValueType.INT64,
                        "featureC": client.ValueType.BOOLEAN,
                        "featureD": client.ValueType.STRING
                    },
                    feature_orders=["featureA", "featureB", "featureC", "featureD"],
                    model_prediction_output=client.ModelPredictionOutput(
                        client.BinaryClassificationOutput(
                            prediction_score_column="prediction_score",
                            actual_score_column="actual_score",
                            positive_class_label="positive",
                            negative_class_label="negative",
                            score_threshold=0.5,
                            output_class=client.ModelPredictionOutputClass.BINARYCLASSIFICATIONOUTPUT
                        )
                    ) 
                )
            ),
            ModelSchema(
                id=1,
                model_id=1,
                spec=InferenceSchema(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags"],
                    feature_types={
                        "featureA": ValueType.FLOAT64,
                        "featureB": ValueType.INT64,
                        "featureC": ValueType.BOOLEAN,
                        "featureD": ValueType.STRING
                    },
                    feature_orders=["featureA", "featureB", "featureC", "featureD"],
                    model_prediction_output=BinaryClassificationOutput(
                        prediction_score_column="prediction_score",
                        actual_score_column="actual_score",
                        positive_class_label="positive",
                        negative_class_label="negative",
                        score_threshold=0.5
                    )
                ),
            ),
            None
        ),
        (
            client.ModelSchema(
                id=2,
                model_id=1,
                spec=client.SchemaSpec(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags", "extras"],
                    feature_types={
                        "featureA": client.ValueType.FLOAT64,
                        "featureB": client.ValueType.INT64,
                        "featureC": client.ValueType.BOOLEAN,
                        "featureD": client.ValueType.STRING
                    },
                    model_prediction_output=client.ModelPredictionOutput(
                        client.RegressionOutput(
                            prediction_score_column="prediction_score",
                            actual_score_column="actual_score",
                            output_class=client.ModelPredictionOutputClass.REGRESSIONOUTPUT
                        )
                    ) 
                )
            ),
            ModelSchema(
                id=2,
                model_id=1,
                spec=InferenceSchema(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags", "extras"],
                    feature_types={
                        "featureA": ValueType.FLOAT64,
                        "featureB": ValueType.INT64,
                        "featureC": ValueType.BOOLEAN,
                        "featureD": ValueType.STRING
                    },
                    model_prediction_output=RegressionOutput(
                        prediction_score_column="prediction_score",
                        actual_score_column="actual_score"
                    )
                ),
            ),
            None
        ),
        (
            client.ModelSchema(
                id=3,
                model_id=1,
                spec=client.SchemaSpec(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags", "extras"],
                    feature_types={
                        "featureA": client.ValueType.FLOAT64,
                        "featureB": client.ValueType.INT64,
                        "featureC": client.ValueType.BOOLEAN,
                        "featureD": client.ValueType.STRING
                    },
                    model_prediction_output=client.ModelPredictionOutput(
                        client.RankingOutput(
                            rank_score_column="score",
                            relevance_score_column="relevance_score",
                            output_class=client.ModelPredictionOutputClass.RANKINGOUTPUT
                        )
                    ) 
                )
            ),
            ModelSchema(
                id=3,
                model_id=1,
                spec=InferenceSchema(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags", "extras"],
                    feature_types={
                        "featureA": ValueType.FLOAT64,
                        "featureB": ValueType.INT64,
                        "featureC": ValueType.BOOLEAN,
                        "featureD": ValueType.STRING
                    },
                    model_prediction_output=RankingOutput(
                        rank_score_column="score",
                        relevance_score_column="relevance_score"
                    )
                ),
            ), 
            None
        ),
         (
            client.ModelSchema(
                id=3,
                model_id=1,
                spec=client.SchemaSpec(
                    session_id_column="session_id",
                    row_id_column="row_id",
                    tag_columns=["tags", "extras"],
                    feature_types={
                        "featureA": client.ValueType.FLOAT64,
                        "featureB": client.ValueType.INT64,
                        "featureC": client.ValueType.BOOLEAN,
                        "featureD": client.ValueType.STRING
                    },
                    model_prediction_output=client.ModelPredictionOutput()
                )
            ),
            None,
            ValueError
        )
    ]
)
def test_model_schema_conversion(response, expected, error):
    if error is None:
        got = ModelSchema.from_model_schema_response(response)
        assert got == expected
        
        client_model_schema = got.to_client_model_schema()
        assert client_model_schema == response
    else:
        with pytest.raises(error):
            ModelSchema.from_model_schema_response(response)
