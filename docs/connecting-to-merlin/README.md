# Connecting to Merlin

## [Python SDK](./python-sdk.md)

- Log Model into Merlin
- Manage Model Version Endpoint and Model Version Endpoint
- Manage Batch Prediction Job

## [Merlin CLI](./merlin-cli.md)

- Deploy the Model Endpoint
- Undeploy the Model Endpoint
- Scaffold a new PyFunc project

## Client Libraries

Merlin provides [Go client library](../../api/client) to deploy and serve ML model in production.

To connect to Merlin deployment, the client needs to be authenticated by Google OAuth2. You can use `google.DefaultClient()` to get the Application Default Credential.

```go
googleClient, _ := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")

cfg := client.NewConfiguration()
cfg.BasePath = "http://merlin.dev/api/merlin/v1"
cfg.HTTPClient = googleClient

apiClient := client.NewAPIClient(cfg)
```
