# Copyright 2020 The Merlin Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import datetime
import json
import types
from unittest.mock import MagicMock, patch

import client
import client as cl
import pytest
from merlin import AutoscalingPolicy, DeploymentMode, MetricsType
from merlin.autoscaling import (
    RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY,
    SERVERLESS_DEFAULT_AUTOSCALING_POLICY,
)
from merlin.batch.config import PredictionJobConfig, ResultType
from merlin.batch.job import JobStatus
from merlin.batch.sink import BigQuerySink, SaveMode
from merlin.batch.source import BigQuerySource
from merlin.endpoint import VersionEndpoint
from merlin.model import ModelType
from merlin.model_schema import InferenceSchema, ModelSchema, RankingOutput, ValueType
from merlin.protocol import Protocol
from merlin.model_observability import ModelObservability
from urllib3_mock import Responses

default_resource_request = cl.ResourceRequest(
    min_replica=1, max_replica=1, cpu_request="100m", memory_request="128Mi"
)
gpu = cl.GPUConfig(
    name="nvidia-tesla-p4",
    values=["1", "4", "8"],
    resource_type="nvidia.com/gpu",
    node_selector={"cloud.google.com/gke-accelerator": "nvidia-tesla-p4"},
    tolerations=[
        cl.GPUToleration(
            key="caraml/nvidia-tesla-p4",
            operator="Equal",
            value="enabled",
            effect="NoSchedule",
        ),
        cl.GPUToleration(
            key="nvidia.com/gpu", operator="Equal", value="present", effect="NoSchedule"
        ),
    ],
)

env_1 = cl.Environment(
    id=1,
    name="dev",
    cluster="cluster-1",
    is_default=True,
    default_resource_request=default_resource_request,
)
env_2 = cl.Environment(
    id=2,
    name="dev-2",
    cluster="cluster-2",
    is_default=False,
    default_resource_request=default_resource_request,
)
env_3 = cl.Environment(
    id=2,
    name="dev-3",
    cluster="cluster-3",
    is_default=False,
    default_resource_request=default_resource_request,
    gpus=[gpu],
)

ep1 = cl.VersionEndpoint(
    id="1234",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_1.name,
    environment=env_1,
    monitoring_url="grafana.com",
)
ep2 = cl.VersionEndpoint(
    id="4567",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-2",
    environment_name=env_2.name,
    environment=env_2,
    monitoring_url="grafana.com",
)
ep3 = cl.VersionEndpoint(
    id="1234",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_1.name,
    environment=env_1,
    monitoring_url="grafana.com",
    deployment_mode="raw_deployment",
)
ep4 = cl.VersionEndpoint(
    id="1234",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_1.name,
    environment=env_1,
    monitoring_url="grafana.com",
    autoscaling_policy=client.AutoscalingPolicy(
        metrics_type=cl.MetricsType.CPU_UTILIZATION, target_value=10
    ),
)
resource_request_with_gpu = cl.ResourceRequest(
    min_replica=1,
    max_replica=1,
    cpu_request="100m",
    memory_request="128Mi",
    gpu_name="nvidia-tesla-p4",
    gpu_request="1",
)
ep5 = cl.VersionEndpoint(
    id="789",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_3.name,
    environment=env_3,
    monitoring_url="grafana.com",
    resource_request=resource_request_with_gpu,
)
upi_ep = cl.VersionEndpoint(
    id="1234",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_1.name,
    environment=env_1,
    monitoring_url="grafana.com",
    protocol=cl.Protocol.UPI_V1,
)
observability_enabled_ep = cl.VersionEndpoint(
    id="7899",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_3.name,
    environment=env_3,
    monitoring_url="grafana.com",
    enable_model_observability=True,
)

more_granular_observability_cfg_ep = cl.VersionEndpoint(
    id="8000",
    version_id=1,
    status="running",
    url="localhost/1",
    service_name="svc-1",
    environment_name=env_3.name,
    environment=env_3,
    monitoring_url="grafana.com",
    model_observability=cl.ModelObservability(
        enabled=True,
        ground_truth_source=cl.GroundTruthSource(
            table_urn="table_urn",
            event_timestamp_column="event_timestamp_column",
            source_project="dwh_project",
        ),
        ground_truth_job=cl.GroundTruthJob(
            cron_schedule="cron_schedule",
            service_account_secret_name="service_account_secret_name",
            start_day_offset_from_now=1,
            end_day_offset_from_now=1,
            cpu_request="cpu_request",
            cpu_limit="cpu_limit",
            memory_request="memory_request",
            memory_limit="memory_limit",
            grace_period_day=1,
        ),
        prediction_log_ingestion_resource_request=cl.PredictionLogIngestionResourceRequest(
            replica=1, cpu_request="1", memory_request="1Gi"
        ),
    ),
)

rule_1 = cl.ModelEndpointRule(
    destinations=[
        cl.ModelEndpointRuleDestination(version_endpoint_id=ep1.id, weight=100)
    ]
)
rule_2 = cl.ModelEndpointRule(
    destinations=[
        cl.ModelEndpointRuleDestination(version_endpoint_id=ep2.id, weight=100)
    ]
)
mdl_endpoint_1 = cl.ModelEndpoint(
    id=1,
    model_id=1,
    model=None,
    status="serving",
    url="localhost/1",
    rule=rule_1,
    environment_name=env_1.name,
    environment=env_1,
)
mdl_endpoint_2 = cl.ModelEndpoint(
    id=2,
    model_id=1,
    model=None,
    status="serving",
    url="localhost/2",
    rule=rule_2,
    environment_name=env_2.name,
    environment=env_2,
)

mdl_endpoint_upi = cl.ModelEndpoint(
    id=1,
    model_id=1,
    model=None,
    status="serving",
    url="localhost/1",
    rule=rule_1,
    environment_name=env_1.name,
    environment=env_1,
    protocol=cl.Protocol.UPI_V1,
)

config = {
    "job_config": {
        "version": "v1",
        "kind": "PredictionJob",
        "name": "job-1",
        "bigquery_source": {
            "table": "project.dataset.source_table",
            "features": ["feature_1", "feature_2"],
            "options": {
                "key": "val",
            },
        },
        "model": {
            "type": "PYFUNC_V2",
            "uri": "gs://my-model/model",
            "result": {"type": "DOUBLE"},
            "options": {
                "key": "val",
            },
        },
        "bigquery_sink": {
            "table": "project.dataset.result_table",
            "staging_bucket": "gs://test",
            "result_column": "prediction",
            "save_mode": 1,
            "options": {
                "key": "val",
            },
        },
    },
    "image_ref": "asia.gcr.io/my-image:1",
    "service_account_name": "my-service-account",
    "resource_request": {
        "driver_cpu_request": "1",
        "driver_memory_request": "1Gi",
        "executor_cpu_request": "1",
        "executor_memory_request": "1Gi",
        "executor_replica": 1,
    },
}
job_1 = cl.PredictionJob(
    id=1,
    name="job-1",
    version_id=1,
    model_id=1,
    config=config,
    status="pending",
    error="",
    created_at="2019-08-29T08:13:12.377Z",
    updated_at="2019-08-29T08:13:12.377Z",
)
job_2 = cl.PredictionJob(
    id=2,
    name="job-2",
    version_id=1,
    model_id=1,
    config=config,
    status="pending",
    error="error",
    created_at="2019-08-29T08:13:12.377Z",
    updated_at="2019-08-29T08:13:12.377Z",
)


def serialize_datetime(obj):
    if isinstance(obj, datetime.datetime):
        return obj.isoformat()
    raise TypeError("Type is not serializable")


class TestProject:
    secret_1 = cl.Secret(id=1, name="secret-1", data="secret-data-1")
    secret_2 = cl.Secret(id=2, name="secret-2", data="secret-data-2")

    def test_create_secret(self, project):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "POST"
            mock_response.status = 200
            mock_response.path = "/v1/projects/1/secrets"
            mock_response.data = json.dumps(self.secret_1.to_dict()).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_request.return_value = mock_response
   
            project.create_secret(self.secret_1.name, self.secret_1.data)
            
            _, last_call_kwargs = mock_request.call_args_list[-1]
            actual_body = json.loads(last_call_kwargs.get("body"))
            
            assert actual_body["name"] == self.secret_1.name
            assert actual_body["data"] == self.secret_1.data

    def test_update_secret(self, project):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/projects/1/secrets"
            mock_response_1.data = json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "PATCH"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/projects/1/secrets/1"
            mock_response_2.data = json.dumps(self.secret_1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]

            project.update_secret(self.secret_1.name, "new-data")

            _, last_call_kwargs = mock_request.call_args_list[-1]
            actual_body = json.loads(last_call_kwargs.get("body"))
            
            assert actual_body["name"] == self.secret_1.name
            assert actual_body["data"] == "new-data"

            mock_response_3 = MagicMock()
            mock_response_3.method = "GET"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/projects/1/secrets"
            mock_response_3.data = json.dumps([self.secret_1.to_dict()]).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            mock_request.side_effect = [mock_response_3]

            with pytest.raises(
                ValueError,
                match=f"unable to find secret {self.secret_2.name} in project {project.name}",
            ):
                project.update_secret(self.secret_2.name, "new-data")
                
    def test_delete_secret(self, project):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/projects/1/secrets"
            mock_response_1.data = json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "DELETE"
            mock_response_2.status = 204
            mock_response_2.path = "/v1/projects/1/secrets/1"
            mock_response_2.data = json.dumps({}).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]

            project.delete_secret(self.secret_1.name)

            mock_response_3 = MagicMock()
            mock_response_3.method = "GET"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/projects/1/secrets"
            mock_response_3.data = json.dumps([self.secret_1.to_dict()]).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            mock_request.side_effect = [mock_response_3]

            with pytest.raises(
                ValueError,
                match=f"unable to find secret {self.secret_2.name} in project {project.name}",
            ):
                project.delete_secret(self.secret_2.name)
                
    def test_list_secret(self, project):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "GET"
            mock_response.status = 200
            mock_response.path = "/v1/projects/1/secrets"
            mock_response.data = json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_request.return_value = mock_response

            secret_names = project.list_secret()
            assert secret_names == [self.secret_1.name, self.secret_2.name]
            
class TestModelVersion:
    def test_list_endpoint(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "GET"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions/1/endpoint"
            mock_response.data = json.dumps([ep1.to_dict(), ep2.to_dict()]).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response

            endpoints = version.list_endpoint()
            assert len(endpoints) == 2
            assert endpoints[0].id == ep1.id
            assert endpoints[1].id == ep2.id
            
    def test_deploy(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(ep1.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(ep1.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep1.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(environment_name=env_1.name)

            assert endpoint.id == ep1.id
            assert endpoint.status.value == ep1.status
            assert endpoint.environment_name == ep1.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.autoscaling_policy == SERVERLESS_DEFAULT_AUTOSCALING_POLICY
            assert endpoint.protocol == Protocol.HTTP_JSON
            
    def test_deploy_upiv1(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(upi_ep.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(upi_ep.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([upi_ep.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(environment_name=env_1.name)

            assert endpoint.id == upi_ep.id
            assert endpoint.status.value == upi_ep.status
            assert endpoint.environment_name == upi_ep.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.autoscaling_policy == SERVERLESS_DEFAULT_AUTOSCALING_POLICY
            assert endpoint.protocol == Protocol.UPI_V1
            
    def test_deploy_using_raw_deployment_mode(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(ep3.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(ep3.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep3.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(
                environment_name=env_1.name, deployment_mode=DeploymentMode.RAW_DEPLOYMENT
            )

            assert endpoint.id == ep3.id
            assert endpoint.status.value == ep3.status
            assert endpoint.environment_name == ep3.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            assert endpoint.deployment_mode == DeploymentMode.RAW_DEPLOYMENT
            assert endpoint.autoscaling_policy == RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY
            
    def test_deploy_with_autoscaling_policy(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(ep4.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(ep4.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep4.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(
                environment_name=env_1.name,
                autoscaling_policy=AutoscalingPolicy(
                    metrics_type=MetricsType.CPU_UTILIZATION, target_value=10
                ),
            )

            assert endpoint.id == ep4.id
            assert endpoint.status.value == ep4.status
            assert endpoint.environment_name == ep4.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.autoscaling_policy.metrics_type == MetricsType.CPU_UTILIZATION
            assert endpoint.autoscaling_policy.target_value == 10
            
    def test_deploy_default_env(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/endpoint"
            mock_response_1.data = json.dumps([]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/environments"
            mock_response_2.data = json.dumps([env_2.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
            
            with pytest.raises(ValueError):
                version.deploy()
                
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(ep1.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(ep1.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep1.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy()

            assert endpoint.id == ep1.id
            assert endpoint.status.value == ep1.status
            assert endpoint.environment_name == ep1.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            
    def test_redeploy_model(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([ep3.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "PUT"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_3.data = json.dumps(ep4.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_4.data = json.dumps(ep4.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep4.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            # Redeployment (add autoscaling policy and change deployment mode)
            endpoint = version.deploy(
                environment_name=env_1.name,
                autoscaling_policy=AutoscalingPolicy(
                    metrics_type=MetricsType.CPU_UTILIZATION, target_value=10
                ),
            )

            assert endpoint.id == ep4.id
            assert endpoint.status.value == ep4.status
            assert endpoint.environment_name == ep4.environment_name
            assert endpoint.environment.cluster == env_1.cluster
            assert endpoint.environment.name == env_1.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.autoscaling_policy.metrics_type == MetricsType.CPU_UTILIZATION
            assert endpoint.autoscaling_policy.target_value == 10
            
    def test_deploy_with_gpu(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_3.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(ep5.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/789"
            mock_response_4.data = json.dumps(ep5.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([ep5.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(environment_name=env_3.name)

            assert endpoint.id == ep5.id
            assert endpoint.status.value == ep5.status
            assert endpoint.environment_name == ep5.environment_name
            assert endpoint.environment.cluster == env_3.cluster
            assert endpoint.environment.name == env_3.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.resource_request.gpu_name == resource_request_with_gpu.gpu_name
            assert (
                endpoint.resource_request.gpu_request
                == resource_request_with_gpu.gpu_request
            )
    
    def test_deploy_with_model_observability_enabled(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_3.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(observability_enabled_ep.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/7899"
            mock_response_4.data = json.dumps(observability_enabled_ep.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([observability_enabled_ep.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            endpoint = version.deploy(
                environment_name=env_3.name, enable_model_observability=True
            )

            assert endpoint.id == observability_enabled_ep.id
            assert endpoint.status.value == observability_enabled_ep.status
            assert endpoint.environment_name == observability_enabled_ep.environment_name
            assert endpoint.environment.cluster == env_3.cluster
            assert endpoint.environment.name == env_3.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.enable_model_observability == True
            
    def test_deploy_with_more_granular_model_observability_cfg(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_3.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/versions/1/endpoint"
            mock_response_3.data = json.dumps(more_granular_observability_cfg_ep.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/endpoint/8000"
            mock_response_4.data = json.dumps(more_granular_observability_cfg_ep.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_5 = MagicMock()
            mock_response_5.method = "GET"
            mock_response_5.status = 200
            mock_response_5.path = "/v1/models/1/versions/1/endpoint"
            mock_response_5.data = json.dumps([more_granular_observability_cfg_ep.to_dict()]).encode('utf-8')
            mock_response_5.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4, mock_response_5]

            model_observability = ModelObservability.from_model_observability_response(more_granular_observability_cfg_ep.model_observability)
            endpoint = version.deploy(
                environment_name=env_3.name, model_observability=model_observability
            )

            assert endpoint.id == more_granular_observability_cfg_ep.id
            assert endpoint.status.value == more_granular_observability_cfg_ep.status
            assert endpoint.environment_name == more_granular_observability_cfg_ep.environment_name
            assert endpoint.environment.cluster == env_3.cluster
            assert endpoint.environment.name == env_3.name
            assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
            assert endpoint.model_observability == model_observability

    def test_undeploy(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "GET"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions/1/endpoint"
            mock_response.data = json.dumps([ep2.to_dict()]).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_request.return_value = mock_response

            version.undeploy(environment_name=env_1.name)
            assert mock_request.call_count == 1
            
            mock_request.reset_mock()
            
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/endpoint"
            mock_response_1.data = json.dumps([ep1.to_dict(), ep2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "DELETE"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_2.data = json.dumps(ep1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]

            version.undeploy(environment_name=env_1.name)
            assert mock_request.call_count == 2
            
    def test_undeploy_default_env(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/endpoint"
            mock_response_1.data = json.dumps([]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/environments"
            mock_response_2.data = json.dumps([env_2.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
            with pytest.raises(ValueError):
                version.deploy()
                
            mock_request.reset_mock()

            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([ep2.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]

            version.undeploy()
            assert mock_request.call_count == 2

            mock_request.reset_mock()
            
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/endpoint"
            mock_response_2.data = json.dumps([ep1.to_dict(), ep2.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "DELETE"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/versions/1/endpoint/1234"
            mock_response_3.data = json.dumps([ep1.to_dict(), ep2.to_dict()]).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3]

            version.undeploy()
            assert mock_request.call_count == 3

    def test_list_prediction_job(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/jobs-by-page?page=1"
            mock_response_1.data = json.dumps({
                "results": [job_1.to_dict()],
                "paging": {
                    "page": 1,
                    "pages": 2,
                    "total": 2,
                },
            }, default=serialize_datetime).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/jobs-by-page?page=2"
            mock_response_2.data = json.dumps({
                "results": [job_2.to_dict()],
                "paging": {
                    "page": 2,
                    "pages": 2,
                    "total": 2,
                },
            }, default=serialize_datetime).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
        
            jobs = version.list_prediction_job()
            assert len(jobs) == 2
            assert jobs[0].id == job_1.id
            assert jobs[0].name == job_1.name
            assert jobs[0].status == JobStatus(job_1.status)
            assert jobs[0].error == job_1.error

            assert jobs[1].id == job_2.id
            assert jobs[1].name == job_2.name
            assert jobs[1].status == JobStatus(job_2.status)
            assert jobs[1].error == job_2.error
            
    def test_create_prediction_job(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            job_1.status = "completed"
            
            mock_response = MagicMock()
            mock_response.method = "POST"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions/1/jobs"
            mock_response.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            bq_src = BigQuerySource(
                table="project.dataset.source_table",
                features=["feature_1", "feature2"],
                options={"key": "val"},
            )

            bq_sink = BigQuerySink(
                table="project.dataset.result_table",
                result_column="prediction",
                save_mode=SaveMode.OVERWRITE,
                staging_bucket="gs://test",
                options={"key": "val"},
            )

            job_config = PredictionJobConfig(
                source=bq_src,
                sink=bq_sink,
                service_account_name="my-service-account",
                result_type=ResultType.INTEGER,
            )
            
            mock_request.return_value = mock_response

            j = version.create_prediction_job(job_config=job_config)
            assert j.status == JobStatus.COMPLETED
            assert j.id == job_1.id
            assert j.error == job_1.error
            assert j.name == job_1.name

            _, last_call_kwargs = mock_request.call_args_list[-1]
            actual_req = json.loads(last_call_kwargs.get("body"))

            assert actual_req["config"]["job_config"]["bigquery_source"] == bq_src.to_dict()
            assert actual_req["config"]["job_config"]["bigquery_sink"] == bq_sink.to_dict()
            assert (
                actual_req["config"]["job_config"]["model"]["result"]["type"]
                == ResultType.INTEGER.value
            )
            assert (
                actual_req["config"]["job_config"]["model"]["uri"]
                == f"{version.artifact_uri}/model"
            )
            assert (
                actual_req["config"]["job_config"]["model"]["type"]
                == ModelType.PYFUNC_V2.value.upper()
            )
            assert actual_req["config"]["service_account_name"] == "my-service-account"
            
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    def test_create_prediction_job_with_retry_failed(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            job_1.status = "pending"

            mock_response = MagicMock()
            mock_response.method = "POST"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions/1/jobs"
            mock_response.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_responses = [mock_response]

            for i in range(5):
                temp_mock_response = MagicMock()
                temp_mock_response.method = "GET"
                temp_mock_response.status = 200
                temp_mock_response.path = "/v1/models/1/versions/1/jobs/1"
                temp_mock_response.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
                temp_mock_response.headers = {
                    'content-type': 'application/json',
                    'charset': 'utf-8'
                }
                
                mock_responses.append(temp_mock_response)

            bq_src = BigQuerySource(
                table="project.dataset.source_table",
                features=["feature_1", "feature2"],
                options={"key": "val"},
            )

            bq_sink = BigQuerySink(
                table="project.dataset.result_table",
                result_column="prediction",
                save_mode=SaveMode.OVERWRITE,
                staging_bucket="gs://test",
                options={"key": "val"},
            )

            job_config = PredictionJobConfig(
                source=bq_src,
                sink=bq_sink,
                service_account_name="my-service-account",
                result_type=ResultType.INTEGER,
            )

            mock_request.side_effect = mock_responses
            with pytest.raises(ValueError):
                j = version.create_prediction_job(job_config=job_config)
                assert j.id == job_1.id
                assert j.error == job_1.error
                assert j.name == job_1.name
                assert mock_request.call_count == 6

    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    def test_create_prediction_job_with_retry_success(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            job_1.status = "pending"

            mock_response_1 = MagicMock()
            mock_response_1.method = "POST"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/jobs"
            mock_response_1.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_responses = [mock_response_1]

            for i in range(4):
                temp_mock_response = MagicMock()
                temp_mock_response.method = "GET"
                temp_mock_response.status = 200
                temp_mock_response.path = "/v1/models/1/versions/1/jobs/1"
                temp_mock_response.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
                temp_mock_response.headers = {
                    'content-type': 'application/json',
                    'charset': 'utf-8'
                }
                
                mock_responses.append(temp_mock_response)

            job_1.status = "completed"
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/jobs/1"
            mock_response_2.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_responses.append(mock_response_2)

            bq_src = BigQuerySource(
                table="project.dataset.source_table",
                features=["feature_1", "feature2"],
                options={"key": "val"},
            )

            bq_sink = BigQuerySink(
                table="project.dataset.result_table",
                result_column="prediction",
                save_mode=SaveMode.OVERWRITE,
                staging_bucket="gs://test",
                options={"key": "val"},
            )

            job_config = PredictionJobConfig(
                source=bq_src,
                sink=bq_sink,
                service_account_name="my-service-account",
                result_type=ResultType.INTEGER,
            )

            mock_request.side_effect = mock_responses
            
            j = version.create_prediction_job(job_config=job_config)
            assert j.status == JobStatus.COMPLETED
            assert j.id == job_1.id
            assert j.error == job_1.error
            assert j.name == job_1.name

            _, last_call_kwargs = mock_request.call_args_list[0]
            actual_req = json.loads(last_call_kwargs.get("body"))
            
            assert actual_req["config"]["job_config"]["bigquery_source"] == bq_src.to_dict()
            assert actual_req["config"]["job_config"]["bigquery_sink"] == bq_sink.to_dict()
            assert (
                actual_req["config"]["job_config"]["model"]["result"]["type"]
                == ResultType.INTEGER.value
            )
            assert (
                actual_req["config"]["job_config"]["model"]["uri"]
                == f"{version.artifact_uri}/model"
            )
            assert (
                actual_req["config"]["job_config"]["model"]["type"]
                == ModelType.PYFUNC_V2.value.upper()
            )
            assert actual_req["config"]["service_account_name"] == "my-service-account"
            assert mock_request.call_count == 6
            
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    def test_create_prediction_job_with_retry_pending_then_failed(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            job_1.status = "pending"

            mock_response_1 = MagicMock()
            mock_response_1.method = "POST"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/jobs"
            mock_response_1.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_responses = [mock_response_1]

            for i in range(4):
                temp_mock_response = MagicMock()
                temp_mock_response.method = "GET"
                temp_mock_response.status = 200
                temp_mock_response.path = "/v1/models/1/versions/1/jobs/1"
                temp_mock_response.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
                temp_mock_response.headers = {
                    'content-type': 'application/json',
                    'charset': 'utf-8'
                }
                
                mock_responses.append(temp_mock_response)

            job_1.status = "failed"
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/jobs/1"
            mock_response_2.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_responses.append(mock_response_2)

            bq_src = BigQuerySource(
                table="project.dataset.source_table",
                features=["feature_1", "feature2"],
                options={"key": "val"},
            )
    
            bq_sink = BigQuerySink(
                table="project.dataset.result_table",
                result_column="prediction",
                save_mode=SaveMode.OVERWRITE,
                staging_bucket="gs://test",
                options={"key": "val"},
            )
    
            job_config = PredictionJobConfig(
                source=bq_src,
                sink=bq_sink,
                service_account_name="my-service-account",
                result_type=ResultType.INTEGER,
            )
    
            mock_request.side_effect = mock_responses
            with pytest.raises(ValueError):
                j = version.create_prediction_job(job_config=job_config)
                assert j.id == job_1.id
                assert j.error == job_1.error
                assert j.name == job_1.name
    
            assert mock_request.call_count == 6
            
    def test_stop_prediction_job(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            job_1.status = "pending"

            mock_response_1 = MagicMock()
            mock_response_1.method = "POST"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions/1/jobs"
            mock_response_1.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions/1/jobs/1"
            mock_response_2.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "PUT"
            mock_response_3.status = 204
            mock_response_3.path = "/v1/models/1/versions/1/jobs/1/stop"
            mock_response_3.data = json.dumps({}).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            job_1.status = "terminated"
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "GET"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/versions/1/jobs/1"
            mock_response_4.data = json.dumps(job_1.to_dict(), default=serialize_datetime).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            bq_src = BigQuerySource(
                table="project.dataset.source_table",
                features=["feature_1", "feature2"],
                options={"key": "val"},
            )

            bq_sink = BigQuerySink(
                table="project.dataset.result_table",
                result_column="prediction",
                save_mode=SaveMode.OVERWRITE,
                staging_bucket="gs://test",
                options={"key": "val"},
            )

            job_config = PredictionJobConfig(
                source=bq_src,
                sink=bq_sink,
                service_account_name="my-service-account",
                result_type=ResultType.INTEGER,
            )
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4]

            j = version.create_prediction_job(job_config=job_config, sync=False)
            j = j.stop()
            assert j.status == JobStatus.TERMINATED
            assert j.id == job_1.id
            assert j.error == job_1.error
            assert j.name == job_1.name
            
    def test_model_version_deletion(self, version):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "DELETE"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions/1"
            mock_response.data = json.dumps(1).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response
            
            response = version.delete_model_version()
            assert response == 1

class TestModel:
    v1 = cl.Version(id=1, model_id=1)
    v2 = cl.Version(id=2, model_id=1)
    v3 = cl.Version(
        id=3, model_id=1, labels={"model": "T-800"}, python_version="3.10.*"
    )

    schema = client.ModelSchema(
        id=3,
        model_id=1,
        spec=client.SchemaSpec(
            tag_columns=["tags", "extras"],
            feature_types={
                "featureA": client.ValueType.FLOAT64,
                "featureB": client.ValueType.INT64,
                "featureC": client.ValueType.BOOLEAN,
                "featureD": client.ValueType.STRING,
            },
            model_prediction_output=client.ModelPredictionOutput(
                client.RankingOutput(
                    rank_score_column="score",
                    relevance_score_column="relevance_score",
                    output_class=client.ModelPredictionOutputClass.RANKINGOUTPUT,
                )
            ),
        ),
    )
    merlin_model_schema = ModelSchema(
        id=3,
        model_id=1,
        spec=InferenceSchema(
            tag_columns=["tags", "extras"],
            feature_types={
                "featureA": ValueType.FLOAT64,
                "featureB": ValueType.INT64,
                "featureC": ValueType.BOOLEAN,
                "featureD": ValueType.STRING,
            },
            model_prediction_output=RankingOutput(
                rank_score_column="score", relevance_score_column="relevance_score"
            ),
        ),
    )
    v4 = cl.Version(
        id=4,
        model_id=1,
        labels={"model": "T-800"},
        python_version="3.10.*",
        model_schema=schema,
    )

    def test_list_version(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/versions?limit=50&cursor=&search="
            mock_response_1.data = json.dumps([self.v1.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8',
                "Next-Cursor": "abcdef"
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/versions?limit=50&cursor=abcdef&search="
            mock_response_2.data = json.dumps([self.v2.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
            
            versions = model.list_version()
            assert len(versions) == 2
            assert versions[0].id == 1
            assert versions[1].id == 2
            
    def test_list_version_with_labels(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "GET"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/versions?limit=50&cursor=&search=labels%3Amodel+in+%28T-800%29"
            mock_response.data = json.dumps([self.v3.to_dict()]).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response
        
            versions = model.list_version({"model": ["T-800"]})
            assert len(versions) == 1
            assert versions[0].id == 3
            assert versions[0].labels["model"] == "T-800"
            
    def test_list_endpoint(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "GET"
            mock_response.status = 200
            mock_response.path = "/v1/models/1/endpoints"
            mock_response.data = json.dumps([mdl_endpoint_1.to_dict(), mdl_endpoint_2.to_dict()]).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response

            endpoints = model.list_endpoint()
            assert len(endpoints) == 2
            assert endpoints[0].id == mdl_endpoint_1.id
            assert endpoints[1].id == mdl_endpoint_2.id
            
    def test_new_model_version(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "POST"
            mock_response.status = 201
            mock_response.path = "/v1/models/1/versions"
            mock_response.data = json.dumps(self.v4.to_dict()).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response

            mv = model.new_model_version(
                labels={"model": "T-800"}, model_schema=self.merlin_model_schema
            )
            assert mv._python_version == "3.10.*"
            assert mv._id == 4
            assert mv._model._id == 1
            assert mv._labels == {"model": "T-800"}
            assert mv._model_schema == self.merlin_model_schema
            
    def test_serve_traffic(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            ve = VersionEndpoint(ep1)
            with pytest.raises(ValueError):
                model.serve_traffic([ve], environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic({ve: -1}, environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic({ve: 101}, environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic(
                    {VersionEndpoint(ep2): 100}, environment_name=env_1.name
                )
                
            # test create
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "POST"
            mock_response_2.status = 201
            mock_response_2.path = "/v1/models/1/endpoints"
            mock_response_2.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            print("aku mock_response_2",mock_response_2)

            mock_request.side_effect = [mock_response_1, mock_response_2]
            
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == mdl_endpoint_1.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )
            assert endpoint.protocol == Protocol.HTTP_JSON
    
            # test update
    
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([mdl_endpoint_1.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/endpoints/1"
            mock_response_2.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "PUT"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/endpoints/1"
            mock_response_3.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3]
            
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == mdl_endpoint_1.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )
            
    def test_stop_serving_traffic(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            ve = VersionEndpoint(ep1)
            with pytest.raises(ValueError):
                model.serve_traffic([ve], environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic({ve: -1}, environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic({ve: 101}, environment_name=env_1.name)

            with pytest.raises(ValueError):
                model.serve_traffic(
                    {VersionEndpoint(ep2): 100}, environment_name=env_1.name
                )
            
            
            # test create
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "POST"
            mock_response_2.status = 201
            mock_response_2.path = "/v1/models/1/endpoints"
            mock_response_2.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
        
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == mdl_endpoint_1.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )
            
            mock_request.reset_mock()

            # test DELETE
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([mdl_endpoint_1.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/endpoints/1"
            mock_response_2.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "DELETE"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/endpoints/1"
            mock_response_3.data = json.dumps({}).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3]
            
            model.stop_serving_traffic(endpoint.environment_name)
            assert mock_request.call_count == 2

    def test_serve_traffic_default_env(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            ve = VersionEndpoint(ep1)
            
            # no default environment
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.return_value = mock_response_1        

            with pytest.raises(ValueError):
                model.serve_traffic({ve: 100})

            # test create
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/endpoints"
            mock_response_2.data = json.dumps([]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "POST"
            mock_response_3.status = 201
            mock_response_3.path = "/v1/models/1/endpoints"
            mock_response_3.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
        
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3]
            
            endpoint = model.serve_traffic({ve: 100})
            assert endpoint.id == mdl_endpoint_1.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )


            # test update
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/environments"
            mock_response_1.data = json.dumps([env_1.to_dict(), env_2.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/endpoints"
            mock_response_2.data = json.dumps([mdl_endpoint_1.to_dict()]).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "GET"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/endpoints/1"
            mock_response_3.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_4 = MagicMock()
            mock_response_4.method = "PUT"
            mock_response_4.status = 200
            mock_response_4.path = "/v1/models/1/endpoints/1"
            mock_response_4.data = json.dumps(mdl_endpoint_1.to_dict()).encode('utf-8')
            mock_response_4.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3, mock_response_4]
        
            endpoint = model.serve_traffic({ve: 100})
            assert endpoint.id == mdl_endpoint_1.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )
            
    def test_serve_traffic_upi(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            ve = VersionEndpoint(upi_ep)
            # test create
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "POST"
            mock_response_2.status = 201
            mock_response_2.path = "/v1/models/1/endpoints"
            mock_response_2.data = json.dumps(mdl_endpoint_upi.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2]
        
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == mdl_endpoint_upi.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
            )
            assert endpoint.protocol == Protocol.UPI_V1

            # test update
            mock_response_1 = MagicMock()
            mock_response_1.method = "GET"
            mock_response_1.status = 200
            mock_response_1.path = "/v1/models/1/endpoints"
            mock_response_1.data = json.dumps([mdl_endpoint_upi.to_dict()]).encode('utf-8')
            mock_response_1.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_2 = MagicMock()
            mock_response_2.method = "GET"
            mock_response_2.status = 200
            mock_response_2.path = "/v1/models/1/endpoints/1"
            mock_response_2.data = json.dumps(mdl_endpoint_upi.to_dict()).encode('utf-8')
            mock_response_2.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_response_3 = MagicMock()
            mock_response_3.method = "PUT"
            mock_response_3.status = 200
            mock_response_3.path = "/v1/models/1/endpoints/1"
            mock_response_3.data = json.dumps(mdl_endpoint_upi.to_dict()).encode('utf-8')
            mock_response_3.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }
            
            mock_request.side_effect = [mock_response_1, mock_response_2, mock_response_3]
        
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == mdl_endpoint_upi.id
            assert (
                endpoint.environment_name == env_1.name == mdl_endpoint_upi.environment_name
            )
            assert endpoint.protocol == Protocol.UPI_V1
            
    def test_model_deletion(self, model):
        with patch("urllib3.PoolManager.request") as mock_request:
            mock_response = MagicMock()
            mock_response.method = "DELETE"
            mock_response.status = 200
            mock_response.path = "/v1/projects/1/models/1"
            mock_response.data = json.dumps(1).encode('utf-8')
            mock_response.headers = {
                'content-type': 'application/json',
                'charset': 'utf-8'
            }

            mock_request.return_value = mock_response
            response = model.delete_model()
    
            assert response == 1