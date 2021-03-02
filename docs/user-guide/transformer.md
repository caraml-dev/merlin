# Transformer

In Merlin ecosystem, Transformer is a service deployed in front of the model service which users can use to perform pre-processing and post-processing steps into the incoming requests before being sent to the model service. The benefits of using Transformer are users can abstract the transformation logic outside of their model and write it in a language more performant than python.

Currently, Merlin has two types of Transformer: Standard and Custom Transformer.

## Standard Transformer

Standard Transformer is a built-in pre and post-processing steps supported by Merlin. Currently, it supports Feast feature enrichment as a pre-processing step.

This feature is available from `merlin-sdk==0.10` and above.

### Deploy Standard Transformer using Merlin UI

Once you logged your model and it’s ready to be deployed, you can go to the model deployment page.

Here’s the short video demonstrating how to configure the Standard Transformer:

![Configure Standard Transformer](configure_standard_transformer.gif)

1. As the name suggests, you must choose **Standard Transformer** as Transformer Type.
2. The **Retrieval Table** panel will be displayed. This panel is where you configure the Feast Project, Entities, and Features to be retrieved.
   1. The list of Feast Entity depends on the selected Feast Project
   2. Similarly, the list of Feast Feature also depends on the configured entities
3. You can have multiple Retrieval Table that can retrieve a different kind of entities and features and enrich the request to your model at once. To add it, simply click `Add Retrieval Table`, and new Retrieval Table panel will be displayed and ready to be configured.
4. You can check the Transformer Configuration YAML specification by clicking `See YAML configuration`. You can copy and paste this YAML and use it for deployment using Merlin SDK.
   1. To read more about Transformer Configuration specification, please continue reading.
5. You can also specify the advanced configuration. These configurations are separated from your model.
   1. Request and response payload logging
   2. Resource request (Replicas, CPU, and memory)
   3. Environment variables (See supported environment variables below)

### Deploy Standard Transformer using Merlin SDK

Make sure you are using the supported version of Merlin SDK.

```bash
> pip install merlin-sdk -U
> pip show merlin-sdk

Name: merlin-sdk
Version: 0.10.0
...
```

You need to pass `transformer` argument to the `merlin.deploy()` function to enable and deploy your standard transformer.

```python
from merlin.resource_request import ResourceRequest
from merlin.transformer import StandardTransformer

# Specify the path to transformer config YAML file
transformer_config_path = "transformer_config.yaml"

# Create the transformer resources requests config
resource_request = ResourceRequest(min_replica=0, max_replica=1,
                                   cpu_request="100m", memory_request="200Mi")

# Create the transformer object
transformer = StandardTransformer(config_file=transformer_config_path,
                                  enabled=True,
                                  resource_request=resource_request)

# Deploy the model alongside the transformer
endpoint = merlin.deploy(v, transformer=transformer)
```

### Standard Transformer Specification

Configuration affecting the Standard Transformer’s retrieval and request enrichment could be configured in the YAML file.

Below is the example (and with description comment):

```yaml
transformerConfig:
  feast:
    - project: default # specify the Feast project where the entities and features reside
      entities: # list of entities to be retrieved
        - name: merchant_id # the entity name
          valueType: STRING # the entity value type. please refer to feast for it
          jsonPath: $.merchants_id # JSON Path syntax on how to parse the entity value from the request
        - name: geohash # the entity name
          valueType: STRING # the entity value type
          udf: Geohash($.bid.from.latitude, $.bid.from.longitude, 7) # using the geohash udf. A list of supported UDFs is documented in later section
      features: # list of features to be retrieved
        - name: merchant_t1_discovery:t1_estimate # feature name. make sure to include feature set
          valueType: DOUBLE # feature value type
          defaultValue: '0' # default value if feature value not exist
```

> To learn more about JSON Path syntax, please go to [this link](https://goessner.net/articles/JsonPath/), [this](https://restfulapi.net/json-jsonpath/) or [this](https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html).

### List of Support UDFs
#### Geohash
Takes in JsonPath to latitude and longitude, and an additional precision parameter to generate the geohash

##### Example

Input
```json
{
  "latitude":1.0,
  "longitude":2.0
}
```

Config
```yaml
- name: geohash
  valueType: STRING
  udf: Geohash($.latitude, $.longitude, 12)
```

Output: `"s01mtw037ms0"`

#### JsonExtract
Used to extract value from JSON string

##### Example

Input
```json
{
  "details": "{\"merchant_id\": 9001}"
}
```

Config
```yaml
- name: merchant_id
  valueType: STRING
  udf: JsonExtract($.details, $.merchant_id)
```

Output: `"9001"`

### Standard Transformer Response Output

The output of the standard transformer becomes the input to your model. Here’s we describe how it looks like so you can consume it correctly in your model.

The response enriches the request by adding a new field (`feast_features`) to the request. The field name is feast_features and it is encoded as Pandas DataFrame compatible encoding with “split” orientation.

For example, the standard transformer receives this JSON payload request:

```json
{
  "merchants": [
    {"id": "1111"},
    {"id": "2222"},
    {"id": "3333"}
  ]
  <other fields>
}
```

After calling Feast and receive the features, the Transformer will enrich the request and send the following enriched request to the model:

```json
{
  "merchants": [
    {"id": "1111"},
    {"id": "2222"},
    {"id": "3333"}
  ],
  <other fields from original request>,
  "feast_features": { # all Feast features are added here
    "merchant_id": { # features are grouped by the entity name
      "columns": [
        "merchant_id",
        "merchant_t1_discovery:t1_estimate",
        "merchant_t2_discovery:t2_discovery",
      ],
      "data": [
        ["1111", 100, 10],
        ["1111", 200, 20],
        ["1111", 300, 30],
      ]
    }
  }
}
```

In your PyFunc model’s infer() function, you can parse the request like this:

```python
def infer():
  merchant_df = pd.DataFrame(**req["feast_features"]["merchant_id"])
```

That logically translates to the following Pandas DataFrame:

merchant_id | merchant_t1_discovery:t1_estimate | merchant_t2_discovery:t2_discovery
------------|-----------------------------------|-----------------------------------
1111 | 100 | 10
2222 | 200 | 20
3333 | 300 | 30

### Standard Transformer Environment Variables

Below are supported environment variables to configure your Transformer.

Name | Description | Default Value
-----|-------------|--------------
`LOG_LEVEL` | Set the logging level for internal system. It doesn’t effect the request-response logging. Supported value: DEBUG, INFO, WARNING, ERROR. | INFO
`FEAST_FEATURE_STATUS_MONITORING_ENABLED` | Enable metrics for the status of each retrieved feature. | false
`FEAST_FEATURE_VALUE_MONITORING_ENABLED` | Enable metrics for the summary value of each retrieved feature. | false

## Custom Transformer

In 0.8 release, Merlin adds support to the Custom Transformer deployment. This transformer type enables the users to deploy their own pre-built Transformer service. The user should develop, build, and publish their own Transformer Docker image.

Similar to Standard Transformer, users can configure Custom Transformer from UI and SDK. The difference is instead of specifying the standard transformer configuration, users configure the Docker image and the command and arguments to run it.

### Deploy Custom Transformer using Merlin UI

1. As the name suggests, you must choose Custom Transformer as Transformer Type.
2. Specify the Docker image registry and name.
   1. You need to push your Docker image into supported registries: public DockerHub repository and private GCR repository.
3. If your Docker image needs command or arguments to start, you can specify them on related input form.
4. You can also specify the advanced configuration. These configurations are separated from your model.
   1. Request and response payload logging
   2. Resource request (Replicas, CPU, and memory)
   3. Environment variables

### Deploy Custom Transformer using Merlin SDK

```python
from merlin.resource_request import ResourceRequest
from merlin.transformer import Transformer

# Create the transformer resources requests config
resource_request = ResourceRequest(min_replica=0, max_replica=1,
                                   cpu_request="100m", memory_request="200Mi")

# Create the transformer object
transformer = Transformer("gcr.io/<your-gcp-project>/<your-docker-image>",
                          resource_request=resource_request)

# Deploy the model alongside the transformer
endpoint = merlin.deploy(v, transformer=transformer)
```
