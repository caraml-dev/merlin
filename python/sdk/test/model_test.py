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

import json
import types
import pytest
from merlin.model import ModelType
from urllib3_mock import Responses
from unittest.mock import patch

import client as cl
from merlin.batch.config import PredictionJobConfig, ResultType
from merlin.batch.job import JobStatus
from merlin.batch.sink import BigQuerySink, SaveMode
from merlin.batch.source import BigQuerySource
from merlin.endpoint import VersionEndpoint

responses = Responses('requests.packages.urllib3')

default_resource_request = cl.ResourceRequest(1, 1, "100m", "128Mi")
env_1 = cl.Environment(1, "dev", "cluster-1", True, default_resource_request=default_resource_request)
env_2 = cl.Environment(2, "dev-2", "cluster-2", False, default_resource_request=default_resource_request)
ep1 = cl.VersionEndpoint("1234", 1, "running", "localhost/1", "svc-1",
                         env_1.name, env_1, "grafana.com")
ep2 = cl.VersionEndpoint("4567", 1, "running", "localhost/1", "svc-2",
                         env_2.name, env_2, "grafana.com")
rule_1 = cl.ModelEndpointRule(destinations=[cl.ModelEndpointRuleDestination(
    ep1.id, weight=100)])
rule_2 = cl.ModelEndpointRule(destinations=[cl.ModelEndpointRuleDestination(
    ep2.id, weight=100)])
mdl_endpoint_1 = cl.ModelEndpoint(1, 1, None, "serving", "localhost/1", rule_1,
                                  env_1.name, env_1)
mdl_endpoint_2 = cl.ModelEndpoint(2, 1, None, "serving", "localhost/2", rule_2,
                                  env_2.name, env_2)

config = {
    "job_config": {
        "version": "v1",
        "kind": "PredictionJob",
        "name": "job-1",
        "bigquery_source": {
            "table": "project.dataset.source_table",
            "features": [
                "feature_1",
                "feature_2"
            ],
            "options": {
                "key": "val",
            }
        },
        "model": {
            "type": "PYFUNC_V2",
            "uri": "gs://my-model/model",
            "result": {
                "type": "DOUBLE"
            },
            "options": {
                "key": "val",
            }
        },
        "bigquery_sink": {
            "table": "project.dataset.result_table",
            "staging_bucket": "gs://test",
            "result_column": "prediction",
            "save_mode": "OVERWRITE",
            "options": {
                "key": "val",
            }
        }
    },
    "image_ref": "asia.gcr.io/my-image:1",
    "service_account_name": "my-service-account",
    "resource_request": {
        "driver_cpu_request": "1",
        "driver_memory_request": "1Gi",
        "executor_cpu_request": "1",
        "executor_memory_request": "1Gi",
        "executor_replica": 1
    }
}
job_1 = cl.PredictionJob(id=1, name="job-1", version_id=1, model_id=1, config=config, status="pending", error="",
                         created_at="2019-08-29T08:13:12.377Z",
                         updated_at="2019-08-29T08:13:12.377Z")
job_2 = cl.PredictionJob(id=2, name="job-2", version_id=1, model_id=1, config=config, status="pending", error="error",
                         created_at="2019-08-29T08:13:12.377Z",
                         updated_at="2019-08-29T08:13:12.377Z")


class TestProject:
    secret_1 = cl.Secret(id=1, name="secret-1", data="secret-data-1")
    secret_2 = cl.Secret(id=2, name="secret-2", data="secret-data-2")

    @responses.activate
    def test_create_secret(self, project):
        responses.add("POST", '/v1/projects/1/secrets',
                      body=json.dumps(self.secret_1.to_dict()),
                      status=200,
                      content_type='application/json')

        project.create_secret(self.secret_1.name, self.secret_1.data)

        actual_body = json.loads(responses.calls[0].request.body)
        assert actual_body["name"] == self.secret_1.name
        assert actual_body["data"] == self.secret_1.data

    @responses.activate
    def test_update_secret(self, project):
        responses.add("GET", '/v1/projects/1/secrets',
                      body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("PATCH", '/v1/projects/1/secrets/1',
                      body=json.dumps(self.secret_1.to_dict()),
                      status=200,
                      content_type='application/json')

        project.update_secret(self.secret_1.name, "new-data")

        actual_body = json.loads(responses.calls[1].request.body)
        assert actual_body["name"] == self.secret_1.name
        assert actual_body["data"] == "new-data"

        responses.reset()

        # test secret not found
        responses.add("GET", '/v1/projects/1/secrets',
                      body=json.dumps([self.secret_1.to_dict()]),
                      status=200,
                      content_type='application/json')

        with pytest.raises(ValueError, match=f"unable to find secret {self.secret_2.name} in project {project.name}"):
            project.update_secret(self.secret_2.name, "new-data")

    @responses.activate
    def test_delete_secret(self, project):
        responses.add("GET", '/v1/projects/1/secrets',
                      body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("DELETE", '/v1/projects/1/secrets/1',
                      status=204,
                      content_type='application/json')

        project.delete_secret(self.secret_1.name)

        responses.reset()

        # test secret not found
        responses.add("GET", '/v1/projects/1/secrets',
                      body=json.dumps([self.secret_1.to_dict()]),
                      status=200,
                      content_type='application/json')

        with pytest.raises(ValueError, match=f"unable to find secret {self.secret_2.name} in project {project.name}"):
            project.delete_secret(self.secret_2.name)

    @responses.activate
    def test_list_secret(self, project):
        responses.add("GET", '/v1/projects/1/secrets',
                      body=json.dumps([self.secret_1.to_dict(), self.secret_2.to_dict()]),
                      status=200,
                      content_type='application/json')

        secret_names = project.list_secret()
        assert secret_names == [self.secret_1.name, self.secret_2.name]

class TestModelVersion:
    @responses.activate
    def test_endpoint(self, version):
        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep2.to_dict()]),
                      status=200,
                      content_type='application/json')

        ep = version.endpoint
        assert ep is None
        responses.reset()

        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
                      status=200,
                      content_type='application/json')

        ep = version.endpoint
        assert ep.id == ep1.id
        assert ep.status.value == ep1.status
        assert ep.environment_name == ep1.environment_name
        assert ep.url.startswith(ep1.url)

    @responses.activate
    def test_list_endpoint(self, version):
        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
                      status=200,
                      content_type='application/json')

        endpoints = version.list_endpoint()
        assert len(endpoints) == 2
        assert endpoints[0].id == ep1.id
        assert endpoints[1].id == ep2.id

    @responses.activate
    def test_deploy(self, version):
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_1.to_dict(), env_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("POST", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps(ep1.to_dict()),
                      status=200,
                      content_type='application/json')
        endpoint = version.deploy(environment_name=env_1.name)

        assert endpoint.id == ep1.id
        assert endpoint.status.value == ep1.status
        assert endpoint.environment_name == ep1.environment_name
        assert endpoint.environment.cluster == env_1.cluster
        assert endpoint.environment.name == env_1.name

    @responses.activate
    def test_deploy_default_env(self, version):
        # no default environment
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        with pytest.raises(ValueError):
            version.deploy()

        # default environment exists
        responses.reset()
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_1.to_dict(), env_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("POST", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps(ep1.to_dict()),
                      status=200,
                      content_type='application/json')
        endpoint = version.deploy()

        assert endpoint.id == ep1.id
        assert endpoint.status.value == ep1.status
        assert endpoint.environment_name == ep1.environment_name
        assert endpoint.environment.cluster == env_1.cluster
        assert endpoint.environment.name == env_1.name

    @responses.activate
    def test_undeploy(self, version):
        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep2.to_dict()]),
                      status=200,
                      content_type='application/json')

        version.undeploy(environment_name=env_1.name)
        assert len(responses.calls) == 1

        responses.reset()
        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("DELETE", '/v1/models/1/versions/1/endpoint/1234',
                      body=json.dumps(ep1.to_dict()),
                      status=200,
                      content_type='application/json')

        version.undeploy(environment_name=env_1.name)
        assert len(responses.calls) == 2

    @responses.activate
    def test_undeploy_default_env(self, version):
        # no default environment
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        with pytest.raises(ValueError):
            version.deploy()

        responses.reset()
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_1.to_dict(), env_2.to_dict()]),
                      status=200,
                      content_type='application/json')

        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep2.to_dict()]),
                      status=200,
                      content_type='application/json')

        version.undeploy()
        assert len(responses.calls) == 2

        responses.reset()
        responses.add("GET", '/v1/environments',
                      body=json.dumps(
                          [env_1.to_dict(), env_2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("GET", '/v1/models/1/versions/1/endpoint',
                      body=json.dumps([ep1.to_dict(), ep2.to_dict()]),
                      status=200,
                      content_type='application/json')
        responses.add("DELETE", '/v1/models/1/versions/1/endpoint/1234',
                      body=json.dumps(ep1.to_dict()),
                      status=200,
                      content_type='application/json')

        version.undeploy()
        assert len(responses.calls) == 3

    @responses.activate
    def test_list_prediction_job(self, version):
        responses.add(method="GET", url='/v1/models/1/versions/1/jobs',
                      body=json.dumps([job_1.to_dict(), job_2.to_dict()]),
                      status=200,
                      content_type='application/json')
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
        responses.add("POST", '/v1/models/1/versions/1/jobs',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        bq_src = BigQuerySource(table="project.dataset.source_table",
                                features=["feature_1", "feature2"],
                                options={"key": "val"})

        bq_sink = BigQuerySink(table="project.dataset.result_table",
                               result_column="prediction",
                               save_mode=SaveMode.OVERWRITE,
                               staging_bucket="gs://test",
                               options={"key": "val"})

        job_config = PredictionJobConfig(source=bq_src,
                                         sink=bq_sink,
                                         service_account_name="my-service-account",
                                         result_type=ResultType.INTEGER)

        j = version.create_prediction_job(job_config=job_config)
        assert j.status == JobStatus.COMPLETED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

        actual_req = json.loads(responses.calls[0].request.body)
        assert actual_req["config"]["job_config"]["bigquery_source"] == bq_src.to_dict()
        assert actual_req["config"]["job_config"]["bigquery_sink"] == bq_sink.to_dict()
        assert actual_req["config"]["job_config"]["model"]["result"]["type"] == ResultType.INTEGER.value
        assert actual_req["config"]["job_config"]["model"]["uri"] == f"{version.artifact_uri}/model"
        assert actual_req["config"]["job_config"]["model"]["type"] == ModelType.PYFUNC_V2.value.upper()
        assert actual_req["config"]["service_account_name"] == "my-service-account"

    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    @responses.activate
    def test_create_prediction_job_with_retry_failed(self, version):
        job_1.status = "pending"
        responses.add("POST", '/v1/models/1/versions/1/jobs',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        for i in range(5):
            responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                          body=json.dumps(job_1.to_dict()),
                          status=500,
                          content_type='application/json')

        bq_src = BigQuerySource(table="project.dataset.source_table",
                                features=["feature_1", "feature2"],
                                options={"key": "val"})

        bq_sink = BigQuerySink(table="project.dataset.result_table",
                               result_column="prediction",
                               save_mode=SaveMode.OVERWRITE,
                               staging_bucket="gs://test",
                               options={"key": "val"})

        job_config = PredictionJobConfig(source=bq_src,
                                         sink=bq_sink,
                                         service_account_name="my-service-account",
                                         result_type=ResultType.INTEGER)

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
        responses.add("POST", '/v1/models/1/versions/1/jobs',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        # Patch the method as currently it is not supported in the library
        # https://github.com/getsentry/responses/issues/135
        def _find_match(self, request):
            for match in self._urls:
                if request.method == match['method'] and \
                        self._has_url_match(match, request.url):
                    return match

        def _find_match_patched(self, request):
            for index, match in enumerate(self._urls):
                if request.method == match['method'] and \
                        self._has_url_match(match, request.url):
                    if request.method == "GET" and request.url == "/v1/models/1/versions/1/jobs/1":
                        return self._urls.pop(index)
                    else:
                        return match
        responses._find_match = types.MethodType(_find_match_patched, responses)

        for i in range(4):
            responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                          body=json.dumps(job_1.to_dict()),
                          status=500,
                          content_type='application/json')

        job_1.status = "completed"
        responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        bq_src = BigQuerySource(table="project.dataset.source_table",
                                features=["feature_1", "feature2"],
                                options={"key": "val"})

        bq_sink = BigQuerySink(table="project.dataset.result_table",
                               result_column="prediction",
                               save_mode=SaveMode.OVERWRITE,
                               staging_bucket="gs://test",
                               options={"key": "val"})

        job_config = PredictionJobConfig(source=bq_src,
                                         sink=bq_sink,
                                         service_account_name="my-service-account",
                                         result_type=ResultType.INTEGER)

        j = version.create_prediction_job(job_config=job_config)
        assert j.status == JobStatus.COMPLETED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

        actual_req = json.loads(responses.calls[0].request.body)
        assert actual_req["config"]["job_config"]["bigquery_source"] == bq_src.to_dict()
        assert actual_req["config"]["job_config"]["bigquery_sink"] == bq_sink.to_dict()
        assert actual_req["config"]["job_config"]["model"]["result"]["type"] == ResultType.INTEGER.value
        assert actual_req["config"]["job_config"]["model"]["uri"] == f"{version.artifact_uri}/model"
        assert actual_req["config"]["job_config"]["model"]["type"] == ModelType.PYFUNC_V2.value.upper()
        assert actual_req["config"]["service_account_name"] == "my-service-account"
        assert len(responses.calls) == 6

        # unpatch
        responses._find_match = types.MethodType(_find_match, responses)

    @patch("merlin.model.DEFAULT_PREDICTION_JOB_DELAY", 0)
    @patch("merlin.model.DEFAULT_PREDICTION_JOB_RETRY_DELAY", 0)
    @responses.activate
    def test_create_prediction_job_with_retry_pending_then_failed(self, version):
        job_1.status = "pending"
        responses.add("POST", '/v1/models/1/versions/1/jobs',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        # Patch the method as currently it is not supported in the library
        # https://github.com/getsentry/responses/issues/135
        def _find_match(self, request):
            for match in self._urls:
                if request.method == match['method'] and \
                        self._has_url_match(match, request.url):
                    return match

        def _find_match_patched(self, request):
            for index, match in enumerate(self._urls):
                if request.method == match['method'] and \
                        self._has_url_match(match, request.url):
                    if request.method == "GET" and request.url == "/v1/models/1/versions/1/jobs/1":
                        return self._urls.pop(index)
                    else:
                        return match
        responses._find_match = types.MethodType(_find_match_patched, responses)

        for i in range(3):
            responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                          body=json.dumps(job_1.to_dict()),
                          status=500,
                          content_type='application/json')

        responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        job_1.status = "failed"
        for i in range(5):
            responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                          body=json.dumps(job_1.to_dict()),
                          status=500,
                          content_type='application/json')

        bq_src = BigQuerySource(table="project.dataset.source_table",
                                features=["feature_1", "feature2"],
                                options={"key": "val"})

        bq_sink = BigQuerySink(table="project.dataset.result_table",
                               result_column="prediction",
                               save_mode=SaveMode.OVERWRITE,
                               staging_bucket="gs://test",
                               options={"key": "val"})

        job_config = PredictionJobConfig(source=bq_src,
                                         sink=bq_sink,
                                         service_account_name="my-service-account",
                                         result_type=ResultType.INTEGER)

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
        responses.add("POST", '/v1/models/1/versions/1/jobs',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        responses.add("PUT", '/v1/models/1/versions/1/jobs/1/stop',
                      status=204,
                      content_type='application/json')

        job_1.status = "terminated"
        responses.add("GET", '/v1/models/1/versions/1/jobs/1',
                      body=json.dumps(job_1.to_dict()),
                      status=200,
                      content_type='application/json')

        bq_src = BigQuerySource(table="project.dataset.source_table",
                                features=["feature_1", "feature2"],
                                options={"key": "val"})

        bq_sink = BigQuerySink(table="project.dataset.result_table",
                               result_column="prediction",
                               save_mode=SaveMode.OVERWRITE,
                               staging_bucket="gs://test",
                               options={"key": "val"})

        job_config = PredictionJobConfig(source=bq_src,
                                         sink=bq_sink,
                                         service_account_name="my-service-account",
                                         result_type=ResultType.INTEGER)

        j = version.create_prediction_job(job_config=job_config, sync=False)
        j = j.stop()
        assert j.status == JobStatus.TERMINATED
        assert j.id == job_1.id
        assert j.error == job_1.error
        assert j.name == job_1.name

class TestModel:
    v1 = cl.Version(id=1, model_id=1)
    v2 = cl.Version(id=2, model_id=1)
    v3 = cl.Version(id=3, model_id=1, labels={"model": "T-800"})

    @responses.activate
    def test_list_version(self, model):
        responses.add("GET", "/v1/models/1/versions?limit=50&cursor=&search=",
                      match_querystring=True,
                      body=json.dumps([self.v1.to_dict()]),
                      status=200,
                      adding_headers={"Next-Cursor": "abcdef"},
                      content_type='application/json')
        responses.add("GET", "/v1/models/1/versions?limit=50&cursor=abcdef&search=",
                      match_querystring=True,
                      body=json.dumps([self.v2.to_dict()]),
                      status=200,
                      content_type='application/json')
        versions = model.list_version()
        assert len(versions) == 2
        assert versions[0].id == 1
        assert versions[1].id == 2

    @responses.activate
    def test_list_version_with_labels(self, model):
        responses.add("GET",
                      "/v1/models/1/versions?limit=50&cursor=&search=labels%3Amodel+in+%28T-800%29",
                      body=json.dumps([self.v3.to_dict()]),
                      status=200,
                      match_querystring=True,
                      content_type='application/json')
        versions = model.list_version({"model": ["T-800"]})
        assert len(versions) == 1
        assert versions[0].id == 3
        assert versions[0].labels["model"] == "T-800"

        @responses.activate
        def test_list_endpoint(self, model):
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps(
                              [mdl_endpoint_1.to_dict(),
                               mdl_endpoint_2.to_dict()]),
                          status=200,
                          content_type='application/json')

            endpoints = model.list_endpoint()
            assert len(endpoints) == 2
            assert endpoints[0].id == str(mdl_endpoint_1.id)
            assert endpoints[1].id == str(mdl_endpoint_2.id)

            v = model.get_version(1)
            assert v.id == 1
            assert model.get_version(3) is None

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
                model.serve_traffic({VersionEndpoint(ep2): 100},
                                    environment_name=env_1.name)

            # test create
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps(
                              []),
                          status=200,
                          content_type='application/json')
            responses.add("POST", '/v1/models/1/endpoints',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == str(mdl_endpoint_1.id)
            assert endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name

            responses.reset()

            # test update
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps([mdl_endpoint_1.to_dict()]),
                          status=200,
                          content_type='application/json')
            responses.add("GET", '/v1/models/1/endpoints/1',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            responses.add("PUT", '/v1/models/1/endpoints/1',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == str(mdl_endpoint_1.id)
            assert endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name

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
                model.serve_traffic({VersionEndpoint(ep2): 100},
                                    environment_name=env_1.name)

            # test create
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps(
                              []),
                          status=200,
                          content_type='application/json')
            responses.add("POST", '/v1/models/1/endpoints',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            endpoint = model.serve_traffic({ve: 100}, environment_name=env_1.name)
            assert endpoint.id == str(mdl_endpoint_1.id)
            assert endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name

            responses.reset()

            # test DELETE
            responses.reset()
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps([mdl_endpoint_1.to_dict()]),
                          status=200,
                          content_type='application/json')
            responses.add("GET", '/v1/models/1/endpoints/1',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            responses.add("DELETE", '/v1/models/1/endpoints/1',
                          status=200,
                          content_type='application/json')
            model.stop_serving_traffic(endpoint.environment_name)
            assert len(responses.calls) == 2

        @responses.activate
        def test_serve_traffic_default_env(self, model):
            ve = VersionEndpoint(ep1)

            # no default environment
            responses.add("GET", '/v1/environments',
                          body=json.dumps(
                              [env_2.to_dict()]),
                          status=200,
                          content_type='application/json')
            with pytest.raises(ValueError):
                model.serve_traffic({ve: 100})

            responses.reset()

            # test create
            responses.add("GET", '/v1/environments',
                          body=json.dumps(
                              [env_1.to_dict(), env_2.to_dict()]),
                          status=200,
                          content_type='application/json')
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps(
                              []),
                          status=200,
                          content_type='application/json')
            responses.add("POST", '/v1/models/1/endpoints',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            endpoint = model.serve_traffic({ve: 100})
            assert endpoint.id == str(mdl_endpoint_1.id)
            assert endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name

            responses.reset()

            # test update
            responses.add("GET", '/v1/environments',
                          body=json.dumps(
                              [env_1.to_dict(), env_2.to_dict()]),
                          status=200,
                          content_type='application/json')
            responses.add("GET", '/v1/models/1/endpoints',
                          body=json.dumps([mdl_endpoint_1.to_dict()]),
                          status=200,
                          content_type='application/json')
            responses.add("GET", '/v1/models/1/endpoints/1',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            responses.add("PUT", '/v1/models/1/endpoints/1',
                          body=json.dumps(mdl_endpoint_1.to_dict()),
                          status=200,
                          content_type='application/json')
            endpoint = model.serve_traffic({ve: 100})
            assert endpoint.id == str(mdl_endpoint_1.id)
            assert endpoint.environment_name == env_1.name == mdl_endpoint_1.environment_name
