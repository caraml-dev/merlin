#!/bin/sh

# Get the value of the MERLIN_PREDICTOR_PORT environment variable
# Get the value of the MERLIN_GUNICORN_WORKERS environment variable
port="$MERLIN_PREDICTOR_PORT"
workers="$WORKERS"

# If the port environment variable is not set, use the default port 8080
# If the workers environment variable is not set, use the default worker number 1

if [ -z "$port" ]; then
    port="8080"
fi

if [ -z "$workers" ]; then
    workers="1"
fi

# Execute the Gunicorn command with the specified port and number of workers
exec uvicorn app:app --host=0.0.0.0 --port=$port --workers=$workers --no-access-log