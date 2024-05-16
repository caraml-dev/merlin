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
from unittest.mock import patch

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

    @responses.activate
    def test_create_secret(self, project):
        responses.add(
            "POST",
            "/v1/projects/1/secrets",
            body=json.dumps(self.secret_1.to_dict()),
            status=200,
            content_type="application/json",
        )

        project.create_secret(self.secret_1.name, self.secret_1.data)
        actual_body = json.loads(responses.calls[0].request.body)
        assert actual_body["name"] == self.secret_1.name
        assert actual_body["data"] == self.secret_1.data

    @responses.activate
    def test_update_secret(self, project):
        responses.add(
            "GET",
            "/v1/projects/1/secrets",
            body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "PATCH",
            "/v1/projects/1/secrets/1",
            body=json.dumps(self.secret_1.to_dict()),
            status=200,
            content_type="application/json",
        )

        project.update_secret(self.secret_1.name, "new-data")

        actual_body = json.loads(responses.calls[1].request.body)
        assert actual_body["name"] == self.secret_1.name
        assert actual_body["data"] == "new-data"

        responses.reset()

        # test secret not found
        responses.add(
            "GET",
            "/v1/projects/1/secrets",
            body=json.dumps([self.secret_1.to_dict()]),
            status=200,
            content_type="application/json",
        )

        with pytest.raises(
            ValueError,
            match=f"unable to find secret {self.secret_2.name} in project {project.name}",
        ):
            project.update_secret(self.secret_2.name, "new-data")

    @responses.activate
    def test_delete_secret(self, project):
        responses.add(
            "GET",
            "/v1/projects/1/secrets",
            body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "DELETE",
            "/v1/projects/1/secrets/1",
            status=204,
            content_type="application/json",
        )

        project.delete_secret(self.secret_1.name)

        responses.reset()

        # test secret not found
        responses.add(
            "GET",
            "/v1/projects/1/secrets",
            body=json.dumps([self.secret_1.to_dict()]),
            status=200,
            content_type="application/json",
        )

        with pytest.raises(
            ValueError,
            match=f"unable to find secret {self.secret_2.name} in project {project.name}",
        ):
            project.delete_secret(self.secret_2.name)

    @responses.activate
    def test_list_secret(self, project):
        responses.add(
            "GET",
            "/v1/projects/1/secrets",
            body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        secret_names = project.list_secret()
        assert secret_names == [self.secret_1.name, self.secret_2.name]


class TestModelVersion:
    @responses.activate
    def test_list_endpoint(self, version):
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        endpoints = version.list_endpoint()
        assert len(endpoints) == 2
        assert endpoints[0].id == ep1.id
        assert endpoints[1].id == ep2.id

    @responses.activate
    def test_deploy(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(ep1.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep1.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep1.to_dict()]),
            status=200,
            content_type="application/json",
        )

        endpoint = version.deploy(environment_name=env_1.name)

        assert endpoint.id == ep1.id
        assert endpoint.status.value == ep1.status
        assert endpoint.environment_name == ep1.environment_name
        assert endpoint.environment.cluster == env_1.cluster
        assert endpoint.environment.name == env_1.name
        assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
        assert endpoint.autoscaling_policy == SERVERLESS_DEFAULT_AUTOSCALING_POLICY
        assert endpoint.protocol == Protocol.HTTP_JSON

    @responses.activate
    def test_deploy_upiv1(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(upi_ep.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(upi_ep.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([upi_ep.to_dict()]),
            status=200,
            content_type="application/json",
        )

        endpoint = version.deploy(environment_name=env_1.name)

        assert endpoint.id == upi_ep.id
        assert endpoint.status.value == upi_ep.status
        assert endpoint.environment_name == upi_ep.environment_name
        assert endpoint.environment.cluster == env_1.cluster
        assert endpoint.environment.name == env_1.name
        assert endpoint.deployment_mode == DeploymentMode.SERVERLESS
        assert endpoint.autoscaling_policy == SERVERLESS_DEFAULT_AUTOSCALING_POLICY
        assert endpoint.protocol == Protocol.UPI_V1

    @responses.activate
    def test_deploy_using_raw_deployment_mode(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(ep3.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep3.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep3.to_dict()]),
            status=200,
            content_type="application/json",
        )

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

    @responses.activate
    def test_deploy_with_autoscaling_policy(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(ep4.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep4.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep4.to_dict()]),
            status=200,
            content_type="application/json",
        )

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

    @responses.activate
    def test_deploy_default_env(self, version):
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        # no default environment
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        with pytest.raises(ValueError):
            version.deploy()

        # default environment exists
        responses.reset()
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(ep1.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep1.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep1.to_dict()]),
            status=200,
            content_type="application/json",
        )

        endpoint = version.deploy()

        assert endpoint.id == ep1.id
        assert endpoint.status.value == ep1.status
        assert endpoint.environment_name == ep1.environment_name
        assert endpoint.environment.cluster == env_1.cluster
        assert endpoint.environment.name == env_1.name

    @responses.activate
    def test_redeploy_model(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep3.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "PUT",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep4.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep4.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep4.to_dict()]),
            status=200,
            content_type="application/json",
        )

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

    @responses.activate
    def test_deploy_with_gpu(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_3.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(ep5.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/789",
            body=json.dumps(ep5.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep5.to_dict()]),
            status=200,
            content_type="application/json",
        )

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

    @responses.activate
    def test_deploy_with_model_observability_enabled(self, version):
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_3.to_dict()]),
            status=200,
            content_type="application/json",
        )
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps(observability_enabled_ep.to_dict()),
            status=201,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([observability_enabled_ep.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint/7899",
            body=json.dumps(observability_enabled_ep.to_dict()),
            status=200,
            content_type="application/json",
        )

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

    @responses.activate
    def test_undeploy(self, version):
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        version.undeploy(environment_name=env_1.name)
        assert len(responses.calls) == 1

        responses.reset()
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "DELETE",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep1.to_dict()),
            status=200,
            content_type="application/json",
        )

        version.undeploy(environment_name=env_1.name)
        assert len(responses.calls) == 2

    @responses.activate
    def test_undeploy_default_env(self, version):
        # This is the additional check which deploy makes to determine if there are any existing endpoints associated
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        # no default environment
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        with pytest.raises(ValueError):
            version.deploy()

        responses.reset()
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        version.undeploy()
        assert len(responses.calls) == 2

        responses.reset()
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions/1/endpoint",
            body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "DELETE",
            "/v1/models/1/versions/1/endpoint/1234",
            body=json.dumps(ep1.to_dict()),
            status=200,
            content_type="application/json",
        )

        version.undeploy()
        assert len(responses.calls) == 3

    @responses.activate
    def test_list_prediction_job(self, version):
        responses.add(
            method="GET",
            url="/v1/models/1/versions/1/jobs-by-page?page=1",
            body=json.dumps({
                "results": [job_1.to_dict()],
                "paging": {
                    "page": 1,
                    "pages": 2,
                    "total": 2,
                },
            }, default=serialize_datetime),
            status=200,
            content_type="application/json",
            match_querystring=True,
        )
        responses.add(
            method="GET",
            url="/v1/models/1/versions/1/jobs-by-page?page=2",
            body=json.dumps({
                "results": [job_2.to_dict()],
                "paging": {
                    "page": 2,
                    "pages": 2,
                    "total": 2,
                },
            }, default=serialize_datetime),
            status=200,
            content_type="application/json",
            match_querystring=True,
        )
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

    @responses.activate
    def test_create_prediction_job(self, version):
        job_1.status = "completed"
        responses.add(
            "POST",
            "/v1/models/1/versions/1/jobs",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

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

        j = version.create_prediction_job(job_config=job_config)
        assert j.status == JobStatus.COMPLETED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

        actual_req = json.loads(responses.calls[0].request.body)
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
    @responses.activate
    def test_create_prediction_job_with_retry_failed(self, version):
        job_1.status = "pending"
        responses.add(
            "POST",
            "/v1/models/1/versions/1/jobs",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

        for i in range(5):
            responses.add(
                "GET",
                "/v1/models/1/versions/1/jobs/1",
                body=json.dumps(job_1.to_dict(), default=serialize_datetime),
                status=500,
                content_type="application/json",
            )

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

        with pytest.raises(ValueError):
            j = version.create_prediction_job(job_config=job_config)
            assert j.id == job_1.id
            assert j.error == job_1.error
            assert j.name == job_1.name
            assert len(responses.calls) == 6

    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    @responses.activate
    def test_create_prediction_job_with_retry_success(self, version):
        job_1.status = "pending"
        responses.add(
            "POST",
            "/v1/models/1/versions/1/jobs",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

        # Patch the method as currently it is not supported in the library
        # https://github.com/getsentry/responses/issues/135
        def _find_match(self, request):
            for match in self._urls:
                if request.method == match["method"] and self._has_url_match(
                    match, request.url
                ):
                    return match

        def _find_match_patched(self, request):
            for index, match in enumerate(self._urls):
                if request.method == match["method"] and self._has_url_match(
                    match, request.url
                ):
                    if (
                        request.method == "GET"
                        and request.url == "/v1/models/1/versions/1/jobs/1"
                    ):
                        return self._urls.pop(index)
                    else:
                        return match

        responses._find_match = types.MethodType(_find_match_patched, responses)

        for i in range(4):
            responses.add(
                "GET",
                "/v1/models/1/versions/1/jobs/1",
                body=json.dumps(job_1.to_dict(), default=serialize_datetime),
                status=500,
                content_type="application/json",
            )

        job_1.status = "completed"
        responses.add(
            "GET",
            "/v1/models/1/versions/1/jobs/1",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

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

        j = version.create_prediction_job(job_config=job_config)
        assert j.status == JobStatus.COMPLETED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

        actual_req = json.loads(responses.calls[0].request.body)
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
        assert len(responses.calls) == 6

        # unpatch
        responses._find_match = types.MethodType(_find_match, responses)

    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    @responses.activate
    def test_create_prediction_job_with_retry_pending_then_failed(self, version):
        job_1.status = "pending"
        responses.add(
            "POST",
            "/v1/models/1/versions/1/jobs",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

        # Patch the method as currently it is not supported in the library
        # https://github.com/getsentry/responses/issues/135
        def _find_match(self, request):
            for match in self._urls:
                if request.method == match["method"] and self._has_url_match(
                    match, request.url
                ):
                    return match

        def _find_match_patched(self, request):
            for index, match in enumerate(self._urls):
                if request.method == match["method"] and self._has_url_match(
                    match, request.url
                ):
                    if (
                        request.method == "GET"
                        and request.url == "/v1/models/1/versions/1/jobs/1"
                    ):
                        return self._urls.pop(index)
                    else:
                        return match

        responses._find_match = types.MethodType(_find_match_patched, responses)

        for i in range(3):
            responses.add(
                "GET",
                "/v1/models/1/versions/1/jobs/1",
                body=json.dumps(job_1.to_dict(), default=serialize_datetime),
                status=500,
                content_type="application/json",
            )

        responses.add(
            "GET",
            "/v1/models/1/versions/1/jobs/1",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

        job_1.status = "failed"
        for i in range(5):
            responses.add(
                "GET",
                "/v1/models/1/versions/1/jobs/1",
                body=json.dumps(job_1.to_dict(), default=serialize_datetime),
                status=500,
                content_type="application/json",
            )

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

        with pytest.raises(ValueError):
            j = version.create_prediction_job(job_config=job_config)
            assert j.id == job_1.id
            assert j.error == job_1.error
            assert j.name == job_1.name

        # unpatch
        responses._find_match = types.MethodType(_find_match, responses)
        assert len(responses.calls) == 10

    @responses.activate
    def test_stop_prediction_job(self, version):
        job_1.status = "pending"
        responses.add(
            "POST",
            "/v1/models/1/versions/1/jobs",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

        responses.add(
            "PUT",
            "/v1/models/1/versions/1/jobs/1/stop",
            status=204,
            content_type="application/json",
        )

        job_1.status = "terminated"
        responses.add(
            "GET",
            "/v1/models/1/versions/1/jobs/1",
            body=json.dumps(job_1.to_dict(), default=serialize_datetime),
            status=200,
            content_type="application/json",
        )

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

        j = version.create_prediction_job(job_config=job_config, sync=False)
        j = j.stop()
        assert j.status == JobStatus.TERMINATED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

    @responses.activate
    def test_model_version_deletion(self, version):
        responses.add(
            "DELETE",
            "/v1/models/1/versions/1",
            body=json.dumps(1),
            status=200,
            content_type="application/json",
        )
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

    @responses.activate
    def test_list_version(self, model):
        responses.add(
            "GET",
            "/v1/models/1/versions?limit=50&cursor=&search=",
            match_querystring=True,
            body=json.dumps([self.v1.to_dict()]),
            status=200,
            adding_headers={"Next-Cursor": "abcdef"},
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/versions?limit=50&cursor=abcdef&search=",
            match_querystring=True,
            body=json.dumps([self.v2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        versions = model.list_version()
        assert len(versions) == 2
        assert versions[0].id == 1
        assert versions[1].id == 2

    @responses.activate
    def test_list_version_with_labels(self, model):
        responses.add(
            "GET",
            "/v1/models/1/versions?limit=50&cursor=&search=labels%3Amodel+in+%28T-800%29",
            body=json.dumps([self.v3.to_dict()]),
            status=200,
            match_querystring=True,
            content_type="application/json",
        )
        versions = model.list_version({"model": ["T-800"]})
        assert len(versions) == 1
        assert versions[0].id == 3
        assert versions[0].labels["model"] == "T-800"

    @responses.activate
    def test_list_endpoint(self, model):
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([mdl_endpoint_1.to_dict(), mdl_endpoint_2.to_dict()]),
            status=200,
            content_type="application/json",
        )

        endpoints = model.list_endpoint()
        assert len(endpoints) == 2
        assert endpoints[0].id == mdl_endpoint_1.id
        assert endpoints[1].id == mdl_endpoint_2.id

    @responses.activate
    def test_new_model_version(self, model):
        responses.add(
            "POST",
            "/v1/models/1/versions",
            body=json.dumps(self.v4.to_dict()),
            status=201,
            content_type="application/json",
        )

        mv = model.new_model_version(
            labels={"model": "T-800"}, model_schema=self.merlin_model_schema
        )
        assert mv._python_version == "3.10.*"
        assert mv._id == 4
        assert mv._model._id == 1
        assert mv._labels == {"model": "T-800"}
        assert mv._model_schema == self.merlin_model_schema

    @responses.activate
    def test_serve_traffic(self, model):
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
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/endpoints",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=201,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
        assert endpoint.id == mdl_endpoint_1.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )
        assert endpoint.protocol == Protocol.HTTP_JSON

        responses.reset()

        # test update
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([mdl_endpoint_1.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "PUT",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=200,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
        assert endpoint.id == mdl_endpoint_1.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )

    @responses.activate
    def test_stop_serving_traffic(self, model):
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
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/endpoints",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=201,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
        assert endpoint.id == mdl_endpoint_1.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )

        responses.reset()

        # test DELETE
        responses.reset()
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([mdl_endpoint_1.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "DELETE",
            "/v1/models/1/endpoints/1",
            status=200,
            content_type="application/json",
        )
        model.stop_serving_traffic(endpoint.environment_name)
        assert len(responses.calls) == 2

    @responses.activate
    def test_serve_traffic_default_env(self, model):
        ve = VersionEndpoint(ep1)

        # no default environment
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        with pytest.raises(ValueError):
            model.serve_traffic({ve: 100})

        responses.reset()

        # test create
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/endpoints",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=201,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100})
        assert endpoint.id == mdl_endpoint_1.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )

        responses.reset()

        # test update
        responses.add(
            "GET",
            "/v1/environments",
            body=json.dumps([env_1.to_dict(), env_2.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([mdl_endpoint_1.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "PUT",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_1.to_dict()),
            status=200,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100})
        assert endpoint.id == mdl_endpoint_1.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )

    @responses.activate
    def test_serve_traffic_upi(self, model):
        ve = VersionEndpoint(upi_ep)
        # test create
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "POST",
            "/v1/models/1/endpoints",
            body=json.dumps(mdl_endpoint_upi.to_dict()),
            status=201,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
        assert endpoint.id == mdl_endpoint_upi.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
        )
        assert endpoint.protocol == Protocol.UPI_V1

        responses.reset()

        # test update
        responses.add(
            "GET",
            "/v1/models/1/endpoints",
            body=json.dumps([mdl_endpoint_upi.to_dict()]),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "GET",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_upi.to_dict()),
            status=200,
            content_type="application/json",
        )
        responses.add(
            "PUT",
            "/v1/models/1/endpoints/1",
            body=json.dumps(mdl_endpoint_upi.to_dict()),
            status=200,
            content_type="application/json",
        )
        endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
        assert endpoint.id == mdl_endpoint_upi.id
        assert (
            endpoint.environment_name == env_1.name == mdl_endpoint_upi.environment_name
        )
        assert endpoint.protocol == Protocol.UPI_V1

    @responses.activate
    def test_model_deletion(self, model):
        responses.add(
            "DELETE",
            "/v1/projects/1/models/1",
            body=json.dumps(1),
            status=200,
            content_type="application/json",
        )
        response = model.delete_model()

        assert response == 1
