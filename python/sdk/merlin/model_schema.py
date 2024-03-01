from __future__ import annotations

from dataclasses import dataclass
from typing import Optional

import client
from dataclasses_json import dataclass_json
from merlin.observability.inference import (
    BinaryClassificationOutput,
    InferenceSchema,
    PredictionOutput,
    RankingOutput,
    RegressionOutput,
    ValueType, DEFAULT_SESSION_ID_COLUMN, DEFAULT_ROW_ID_COLUMN,
)
from merlin.util import autostr, extract_optional_value_with_default

@autostr
@dataclass_json
@dataclass
class ModelSchema:
    """
    Representation of schema for a model
    """

    spec: InferenceSchema
    id: Optional[int] = None
    model_id: Optional[int] = None

    @classmethod
    def from_model_schema_response(
        cls, response: Optional[client.ModelSchema] = None
    ) -> Optional[ModelSchema]:
        """
        Convert model schema payload from server response and convert it to `ModelSchema`

        :param response: Model schema payload as part of response from server that already deserialize to OpenAPI ModelSchema
        :type response: Optional[client.ModelSchema]
        """
        if response is None:
            return None

        response_spec = response.spec
        if response_spec is None:
            return ModelSchema(id=response.id, model_id=response.model_id)

        prediction_output = cls.model_prediction_output_from_response(
            response_spec.model_prediction_output
        )
        feature_types = {}
        for key, val in response_spec.feature_types.items():
            feature_types[key] = ValueType(val.value)

        return ModelSchema(
            id=response.id,
            model_id=response.model_id,
            spec=InferenceSchema(
                feature_types=feature_types,
                feature_orders=response_spec.feature_orders,
                session_id_column=extract_optional_value_with_default(response_spec.session_id_column, DEFAULT_SESSION_ID_COLUMN),
                row_id_column=extract_optional_value_with_default(response_spec.row_id_column, DEFAULT_ROW_ID_COLUMN),
                tag_columns=response_spec.tag_columns,
                model_prediction_output=prediction_output,
            ),
        )

    @classmethod
    def model_prediction_output_from_response(
        cls, model_prediction_output: client.ModelPredictionOutput
    ) -> PredictionOutput:
        """
        Convert model prediction output from server payload into `PredictionOutput`.

        :param model_prediction_output: Model prediction output information that is part of model schema server payload.
        :type model_prediction_output: client.ModelPredictionOutput
        """
        actual_instance = model_prediction_output.actual_instance
        if isinstance(actual_instance, client.BinaryClassificationOutput):
            return BinaryClassificationOutput(
                prediction_score_column=actual_instance.prediction_score_column,
                actual_score_column=actual_instance.actual_score_column,
                positive_class_label=actual_instance.positive_class_label,
                negative_class_label=actual_instance.negative_class_label,
                score_threshold=extract_optional_value_with_default(
                    actual_instance.score_threshold, 0.5
                ),
            )
        elif isinstance(actual_instance, client.RegressionOutput):
            return RegressionOutput(
                actual_score_column=actual_instance.actual_score_column,
                prediction_score_column=actual_instance.prediction_score_column,
            )
        elif isinstance(actual_instance, client.RankingOutput):
            return RankingOutput(
                relevance_score_column=extract_optional_value_with_default(
                    actual_instance.relevance_score_column, ""
                ),
                rank_score_column=actual_instance.rank_score_column,
            )
        raise ValueError(
            "model prediction output from server is not in acceptable type"
        )

    def _to_client_prediction_output_spec(self) -> client.ModelPredictionOutput:
        prediction_output = self.spec.model_prediction_output
        if isinstance(prediction_output, BinaryClassificationOutput):
            return client.ModelPredictionOutput(
                client.BinaryClassificationOutput(
                    prediction_score_column=prediction_output.prediction_score_column,
                    actual_score_column=prediction_output.actual_score_column,
                    positive_class_label=prediction_output.positive_class_label,
                    negative_class_label=prediction_output.negative_class_label,
                    score_threshold=prediction_output.score_threshold,
                    output_class=client.ModelPredictionOutputClass(
                        BinaryClassificationOutput.__name__
                    ),
                )
            )
        elif isinstance(prediction_output, RegressionOutput):
            return client.ModelPredictionOutput(
                client.RegressionOutput(
                    actual_score_column=prediction_output.actual_score_column,
                    prediction_score_column=prediction_output.prediction_score_column,
                    output_class=client.ModelPredictionOutputClass(
                        RegressionOutput.__name__
                    ),
                )
            )
        elif isinstance(prediction_output, RankingOutput):
            return client.ModelPredictionOutput(
                client.RankingOutput(
                    relevance_score_column=prediction_output.relevance_score_column,
                    rank_score_column=prediction_output.rank_score_column,
                    output_class=client.ModelPredictionOutputClass(
                        RankingOutput.__name__
                    ),
                )
            )

        raise ValueError("model prediction output is not recognized")

    def to_client_model_schema(self) -> client.ModelSchema:
        """
        Convert `ModelSchema` into OpenAPI `ModelSchema` that is being used by SDK to communicate to merlin server
        """
        feature_types = {}
        for key, val in self.spec.feature_types.items():
            feature_types[key] = client.ValueType(val.value)

        return client.ModelSchema(
            id=self.id,
            model_id=self.model_id,
            spec=client.SchemaSpec(
                session_id_column=self.spec.session_id_column,
                row_id_column=self.spec.row_id_column,
                tag_columns=self.spec.tag_columns,
                feature_types=feature_types,
                feature_orders=self.spec.feature_orders,
                model_prediction_output=self._to_client_prediction_output_spec(),
            ),
        )
