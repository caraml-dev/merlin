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

responses = Responses("requests.packages.urllib3")

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
            
            print(mock_response_1.data)
            
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