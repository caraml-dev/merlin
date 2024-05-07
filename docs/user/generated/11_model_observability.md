<!-- page-title: Model Observability -->
# Model Observability
Model observability enable model's owners to observe and analyze their model in production by looking at the performance and drift metrics. Performance indicate how well your model to do prediction compare to the actual output, and drift indicate the difference of distribution between two datasets. To calculate those metrics the model observability system needs the following data:
* Features data. The features data that is supplied to the model to do prediction
* Prediction data. The prediction as output of your model
* Ground truth / Actual data. The actual value of thing that your model try to predict

Those data can be collected from training phase and serving phase (production). Data that is collected on the training phase is used as the baseline dataset, we can refer it as training dataset. For data during serving phase we can refer it as production dataset, this data must be emitted by the model. By default the merlin model is not emitting any of those data, hence model observability is not enabled by default. However, merlin provides a way so model can emit such data but currently it is limited only for PyFunc model. The way is to turn on the flag of `ENABLE_MODEL_OBSERVABILITY` and modify the PyFunc model to returning model input (features) and model output (prediction output), more detail will be explained in the `Onboarding` section.

## Architecture

![architecture](../../diagrams/model_observability.drawio.svg)

From above architecture diagram, we can see that there are three places where the data is published to model observability system
* Model deployment workflow. Especially after model training step is completed. This step is publishing training dataset as baseline dataset
* Model serving. PyFunc model will emit features and predictions data to a topic in a kafka cluster, and separate kafka consumer consume corresponding topic and publish the data to model observability system. Kafka consumer also store the data into separate BQ table that later will be used to be joined with user ground truth BQ table.
* Ground truth collector workflow. This workflow primary objective is to publish ground truth or actual for a prediction

## Onboarding
As the architecture diagram illustrate, the end to end model onboarding to model observability needs to involving several components. The scope of this section is limited to merlin model modification. 

### PyFunc modification
PyFunc model should implements class `PyFuncV3Model` instead of `PyFuncModel`. This `PyFuncV3Model` has difference method signature that must be implemented. Following are the new methods:

| Method Name | Description |
|-------------|-------------|
| `preprocess(self, request: dict, **kwargs) -> ModelInput` | Doing preprocessing that returning all the required features for prediction. Must be implemented if using `HTTP_JSON` protocol |
| `postprocess(self, model_output: ModelOutput, request: dict) -> dict` | Postprocessing basically do additional processing to construct end result of the overall model. Must be implemented if using `HTTP_JSON` protocol |
| `upiv1_preprocess(self, request: upi_pb2.PredictValuesRequest, context: grpc.ServicerContext) -> ModelInput` | Preprocess method signature that only called when using `UPI_V1` protocol. Must be implemented if using `UPI_V1` protocol |
| `upiv1_postprocess(self, model_output: ModelOutput, request: upi_pb2.PredictValuesRequest) -> upi_pb2.PredictValuesResponse` | Postprocess method signature that only callend when using `UPI_V1` protocol. Must be implemented if using `UPI_V1` protocol |

Beside changes in signature, you can see some of those methods returning new type, `ModelInput` and `ModelOutput`. `ModelInput` is a class that represents input information of the models, this class contains following fields:

| Field | Type | Description|
|-------|------|------------|
| `prediction_ids` | List[str] | Unique identifier for each row in prediction |
| `features` | Union[Values, pandas.DataFrame] | Features value that is used by the model to generate prediction. Length of features should be the same with `prediction_ids` | 
| `entities` | Optional[Union[Values, pandas.DataFrame]] | Additional data that are not used for prediction, but this data is used to retrieved another features, e.g `driver_id`, we can retrieve features associated with certain `driver_id`|
| `session_id` | str | Identifier for the request. This value will be used together with `prediction_ids` as prediction identifier in model observability system  |

`ModelInput` data is essential for model observability since it contains features values and identifier of prediction. Features values are used to calculate feature drift, and identifier is used as join key between features, prediction data with ground truth data. On the other hand, `ModelOutput` is the class that represent raw model prediction output, not the final output of PyFunc model. `ModelOutput` class contains following fields:

| Field | Type | Description |
|-------|------|-------------|
| `prediction` | Values | `predictions` contains prediction output from ml_predict, it may contains multiple columns e.g for multiclass classification or for binary classification that contains prediction score and label |
| `prediction_ids` | List[str] | Unique identifier for each row in prediction output |

Same like `ModelInput`, `ModelOutput` is also essential for model observability, it can be used to calculate prediction drift but more importantly it can calculate performance metrics.

### Standard Model
Only xgboost model is supported at the moment. To enable model observability, user will need to enable request response logging for the model. 1% of the predictions will be sampled, and this is not configurable at the moment.

### Configure Model Schema

Model schema is essential for model observability because it is used by the kafka consumer to choose which columns that is relevant to model observability and do necessary preprocessing before publishing the data to model observability system. Users can see more detail of configuring model schema [here](../templates/09_model_schema.md)

### Deployment
There is not much change on the deployment part, users just needs to set `enable_model_observability` parameter to `True` during model deploy. For clarity, we take one use case for model observability example, suppose a model has 4 features:
* featureA that has float type
* featureB that has int type
* featureC that has string type
* featureD that has float type

The model type is ranking with prediction group id information is located in `session_id` column, prediction id in `prediction_id` column, rank score in `score` column and `relevance_score_column` in `relevance_score`. Below is the snipped of the python code

```python
class ModelObservabilityModel(PyFuncV3Model):

    def preprocess(self, request: dict, **kwargs) -> ModelInput:
        return ModelInput(
            session_id="session_id",
            prediction_ids=["prediction_1", "prediction_2"],
            features=pd.DataFrame([[0.7, 200, "ID", True], [0.99, 250, "SG", False]], columns=["featureA", "featureB", "featureC", "featureD"]),
        )

    def infer(self, model_input: ModelInput) -> ModelOutput:
        return ModelOutput(
            prediction_ids=model_input.prediction_ids,
            predictions=Values(columns=["score"], data=[[0.5], [0.9]]),
        )
    def postprocess(self, model_output: ModelOutput, request: dict) -> dict:
        return {"predictions": model_output.predictions.data}


model_schema = ModelSchema(spec=InferenceSchema(
        feature_types={
            "featureA": ValueType.FLOAT64,
            "featureB": ValueType.INT64,
            "featureC": ValueType.STRING,
            "featureD": ValueType.BOOLEAN
        },
        session_id_column="session_id",
        row_id_column="prediction_id",
        model_prediction_output=RankingOutput(
            rank_score_column="score",
            prediction_group_id_column="session_id",
            relevance_score_column="relevance_score"
        )
    ))
with merlin.new_model_version(model_schema=model_schema) as v:
        v.log_pyfunc_model(model_instance=ModelObservabilityModel(),
                           conda_env="env.yaml",
                           code_dir=["src"],
                           artifacts={"model": ARTIFACT_PATH})
endpoint = merlin.deploy(v, enable_model_observability=True)
```

For standard model, we will also need to add an additional field, `feature_orders` to the inference schema:

```python
InferenceSchema(
    ...
    feature_orders=["featureA", "featureB", "featureC", "featureD"],
)
```

The features in the list should follow the same order as the expected input to the standard model.

The json request body to the standard model is also expected to have these additional fields:
* `session_id (String)` - Identifier for the request.
* `row_ids (Array)` - Unique identifier for each row in prediction. Optional if you only have one row per prediction.