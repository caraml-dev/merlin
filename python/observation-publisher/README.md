# Observation Publisher
Observation publisher consumes observation logs published by Merlin models, then publishes the logs to an observability backend. The backend will then compute the different metrics required to monitor the model's performance and feature distribution. 

## Usage

### Configuration
Environment configuration should be placed under conf/environment.
Refer to example-overrides.yaml for documentation on available configurations.

### Run consumer

```bash
export ENVIRONMENT_CONFIG=<config file name without .yaml extension>
make run
```


## Development

### Run test
```bash
make test
```

### Building Docker Image

```bash
make image
```
