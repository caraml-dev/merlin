<!-- page-title: Webhook -->

# Webhook

When an event happens (e.g. creating new model's version, deploy endpoint), we might want to trigger action in other services, Merlin provides a way this by triggering a webhook call. 
The webhook that will be called is configured per event type, and it could contain multiple endpoints which can be configured as synchronous or asynchronous call. 

For the detail of available configuration please refer to [MLP library docs](https://github.com/caraml-dev/mlp/blob/main/api/pkg/webhooks/README.md), this documentation will only cover the basic config and the payload of webhook which we have implemented.

## 1. Webhook Configuration 
In Merlin's yaml for config you can add new field for the webhook's configuration which will looks like this:
```yaml
WebhooksConfig:
  Enabled: true
  Config:
    on-model-version-predeployment:
      - URL: http://127.0.0.1:8000/async-webhook
        Method: POST
        Async: true
        Name: async
    on-model-version-deployed:
      - URL: http://127.0.0.1:8000/async-webhook
        Method: POST
        Async: true
        Name: async
    on-model-version-undeployed:
      - URL: http://127.0.0.1:8000/sync-webhook
        Method: POST
        FinalResponse: true
        Name: sync
```
## 2. Webhook Events and Payload
### 2.1. Model Version's Endpoint
**Events:**
- `on-model-version-predeployment`: will be triggered before the version endpoint is deployed, we will only proceed with the deployment if the webhook return a `200` response, otherwise if the webhook return is a non `200` response, the deployment will be **aborted**. In both case, we will not take into account of the response body, we will only look at the response code.
- `on-model-version-deployed`: will be triggered after the version endpoint is successfully deployed (either new or updated), there will be no side effect if the webhook return a success or error response, we will only log the response if an error occurred.
- `on-model-version-undeployed`: will be triggered after the version endpoint is successfully undeployed, there will be no side effect if the webhook return a success or error response, we will only log the response if an error occurred.

**Request Payload**
```json
{
  "event_type": "{event type}",
  "version_endpoint": "{version endpoint object}"
}
```

Examples:
```json
{
  "event_type": "on-model-version-predeployment",
  "version_endpoint": {
    "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
    "version_id": 0,
    "status": "pending",
    ...
  }
}
```

For the complete field of `version_endpoint` please refer to our [swagger docs](../../../../swagger.yaml).

**Response** 

As of writing this document (25 September 2024) and mentioned previously, the successful response will not affect the process of our deployment. But for future usage, we're expecting the response of the final synchronous webhook that will be called to be in the following format:

Successful Response:
```json
{
  "code": "200",
  "version_endpoint": "{version endpoint object}"
}
```

Error response:
```json
{
  "code": "{error code}",
  "message": "{reason for error}"
}
```