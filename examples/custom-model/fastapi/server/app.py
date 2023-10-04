import os
from fastapi import FastAPI
import xgboost as xgb

app = FastAPI()

# set global env
MERLIN_MODEL_NAME = os.environ.get("MERLIN_MODEL_NAME", "my-model")
MERLIN_PREDICTOR_PORT = int(os.environ.get("MERLIN_PREDICTOR_PORT", 8080))
MERLIN_ARTIFACT_LOCATION = os.environ.get("MERLIN_ARTIFACT_LOCATION", "model")

# Set API endpoints
API_HEALTH_ENDPOINT = "/"
API_ENDPOINT = f"/v1/models/{MERLIN_MODEL_NAME}"
API_ENDPOINT_PREDICT = f"{API_ENDPOINT}:predict"

# Print Endpoint Info
print("Starting API server")
print(f"Starting API server: {API_ENDPOINT}")
print(f"Starting API predict server: {API_ENDPOINT_PREDICT}")
print(f"Artifact Location : {MERLIN_ARTIFACT_LOCATION}")

# Create Prediction Class
class XgbModel:
    def __init__(self):
        self.loaded = False
        self.model = xgb.Booster({'nthread': 4})
        self.load_model(MERLIN_ARTIFACT_LOCATION)

    def load_model(self, model_path):
        model_file = os.path.join((model_path), 'model.bst')
        self.model.load_model(model_file)
        self.loaded = True

    def predict(self, request):
        data = request['instances']
        dmatrix = xgb.DMatrix(data)
        predictions = self.model.predict(dmatrix)
        return {"response": predictions.tolist(), "status": "ok"}

# Init Class
prediction_model = XgbModel()


# API Endpoints
@app.get(API_HEALTH_ENDPOINT)
async def root():
    return {"message": "API Ready"}

@app.get(API_ENDPOINT)
async def predict_status():
    if prediction_model.loaded:
        return {"message": "Model is loaded"}
    else:
        return {"message": "Model is not loaded"}, 503

@app.post(API_ENDPOINT_PREDICT)
async def predict(request:dict):
    response = prediction_model.predict(request)
    return response
