# Deploy PyFunc Model

Try out the notebooks below to learn how to deploy PyFunc Models to Merlin.

**Note on compatibility**: The Pyfunc servers are compatible with `protobuf>=3.12.0,<5.0.0`. Users whose models have a strong dependency on Protobuf `3.x.x` are advised to pin the library version in their conda environment, when submitting the model version. If using Protobuf `3.x.x`, users can do one of the following:
* Use `protobuf>=3.20.0` - these versions support simplified class definitions and this is the recommended approach.
* If you must use `protobuf>=3.12.0,<3.20.0`, other packages used in the Pyfunc server need to be downgraded as well. Please pin the following in your model’s conda environment:
```yaml
dependencies:
  - pip:
    - protobuf==3.15.6 # Example older protobuf version
    - caraml-upi-protos<=0.3.6
    - grpcio<1.49.0
    - grpcio-reflection<1.49.0
    - grpcio-health-checking<1.49.0
```

## Deploy PyFunc Model

{% embed url="https://github.com/caraml-dev/merlin/blob/main/examples/pyfunc/Pyfunc.ipynb" %}

## Deploy PyFunc Model with Custom Prometheus Metrics

{% embed url="https://github.com/caraml-dev/merlin/blob/main/examples/metrics/Metrics.ipynb" %}