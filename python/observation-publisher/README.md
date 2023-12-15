# Observation Publisher
Observation publisher consumes prediction logs published by Merlin models, then publishes the logs to an observability backend. The backend will then compute the different metrics required to monitor the model's performance and feature distribution. 

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
### Setup
```bash
make setup
```

### Updating requirements.txt
Make changes on requirements.in, then execute
```bash
make pip-compile
```

### Run test
```bash
make test
```

### Building Docker Image
The docker build context is <repository-root>/python. From the build context directory, execute

```bash
make observation-publisher
```
