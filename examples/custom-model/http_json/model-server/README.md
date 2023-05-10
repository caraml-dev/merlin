# Merlin Custom Model
This is sample of web service that will be used as custom model in merlin. The web service will has several endpoints, below are the list:
* **GET** `/` . Server liveness API, return 200 if server is alive (**REQUIRED**)
* **GET** `/v1/models/{model_name}` . Model Health API. Model can receive traffic after this endpoint returning 200 status code (**REQUIRED**)
* **POST** `/v1/models/{model_name}:predict`. Prediction API. This is the endpoint that handle any inference request (**REQUIRED**)
* **GET** `/metrics`. Prometheus will scrape this endpoint to get all the metrics produce by the server (**OPTIONAL**)


## Web Service Environment Variable
Merlin will provide environment variables, which values must be used in the web service. Below are the list of variables:
* `MERLIN_MODEL_NAME`. The value will be used as endpoint path /v1/models/{model_name}:predict. Merlin set this value
* `MERLIN_PREDICTOR_PORT`. Web service must open and listen to this value
* `MERLIN_ARTIFACT_LOCATION`. Local path where artifacts is stored

Users can specify their own variables depending on the web service. For this sample users specify `MODEL_FILE_NAME` variables. The web service uses this variable to know the name of xgboost artifacts.

### Requirements
* Go 1.16
* Docker

### Getting Started
* When running in local you should set above environment variables
 ```
    export MERLIN_MODEL_NAME="sample"
    export MERLIN_PREDICTOR_PORT="8080"
    export MERLIN_ARTIFACT_LOCATION="../xgboost-model"
    export MODEL_FILE_NAME="model.bst"
 ```
* Build go web service
 ```
    go build -o bin/custom_model
 ```
* Run web service
 ```
    ./bin/custom_model
 ```
* Verify web service is running
 ```
    curl --request POST 'http://localhost:8080/v1/models/sample:predict' \
    --header 'Content-Type: application/json' \
    --data '{
        "instances": [
            [2.8,  1.0,  6.8,  0.4],
            [3.1,  1.4,  3.5,  1.6],
            [3.1,  5.4,  0.5,  9.0]
        ]
    }'
 ```