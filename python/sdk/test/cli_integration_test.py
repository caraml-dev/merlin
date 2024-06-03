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

import os
import traceback
from test.utils import undeploy_all_version

import merlin
import pytest
from click.testing import CliRunner
from merlin.merlin import cli
from merlin.model import ModelType


@pytest.fixture()
def deployment_info():
    dirname = os.path.dirname(os.path.dirname(__file__))
    filename = os.path.join(dirname, 'test/sklearn-model')
    url = os.environ.get("E2E_MERLIN_URL", default="http://127.0.0.1:8080")
    project = os.environ.get("E2E_PROJECT_NAME", default="integration-test")
    env = os.environ.get("E2E_MERLIN_ENVIRONMENT", default="id-dev")

    info = {
        'env': env,
        'model_dir': filename,
        'model_type': "sklearn",
        'project': project,
        'url': url,
        'min_replica': '1',
        'max_replica': '1',
        'cpu_request': '100m',
        'cpu_limit': '2',
        'memory_request': '128Mi',
    }
    return info


@pytest.fixture(scope="module")
def runner():
    return CliRunner()


@pytest.mark.cli
@pytest.mark.integration
def test_cli_deployment_undeployment(deployment_info, runner, use_google_oauth):

    model_name = 'cli-test'
    merlin.set_url(deployment_info['url'], use_google_oauth=use_google_oauth)
    merlin.set_project(deployment_info['project'])
    merlin.set_model(model_name, ModelType.SKLEARN)

    undeploy_all_version()

    # Deployment
    result = runner.invoke(
        cli, [
            'deploy',
            '--env', deployment_info['env'],
            '--model-type', deployment_info['model_type'],
            '--model-dir', deployment_info['model_dir'],
            '--model-name', model_name,
            '--project', deployment_info['project'],
            '--url', deployment_info['url']
            ]
        )

    if result.exception:
        traceback.print_exception(*result.exc_info)

    test_deployed_model_version = result.output.split('\n')[0].split(' ')[-1]

    # Get latest deployed model's version
    merlin.set_url(deployment_info['url'], use_google_oauth=use_google_oauth)
    merlin.set_project(deployment_info['project'])
    merlin.set_model(model_name, ModelType.SKLEARN)

    merlin_active_model = merlin.active_model()
    all_versions = merlin_active_model.list_version()

    latest_version = all_versions[0]

    # Undeployment
    undeploy_result = runner.invoke(
        cli, [
            'undeploy',
            '--model-version', test_deployed_model_version,
            '--model-name', model_name,
            '--project', deployment_info['project'],
            '--url', deployment_info['url']
            ]
        )
    if result.exception:
        traceback.print_exception(*result.exc_info)

    planned_output = "Deleting deployment of model {} version {}".format(
        model_name,
        test_deployed_model_version
    )
    received_output = undeploy_result.output.split(' from')[0]

    assert latest_version._id == int(test_deployed_model_version)
    assert received_output == planned_output


@pytest.mark.cli
@pytest.mark.integration
def test_cli_deployment_undeployment_with_resource_request(deployment_info, runner, use_google_oauth):

    model_name = 'cli-resource-request-test'
    merlin.set_url(deployment_info['url'], use_google_oauth=use_google_oauth)
    merlin.set_project(deployment_info['project'])
    merlin.set_model(model_name, ModelType.SKLEARN)

    undeploy_all_version()

    # Deployment
    result = runner.invoke(
        cli, [
            'deploy',
            '--env', deployment_info['env'],
            '--model-type', deployment_info['model_type'],
            '--model-dir', deployment_info['model_dir'],
            '--model-name', model_name,
            '--project', deployment_info['project'],
            '--url', deployment_info['url'],
            '--min-replica', deployment_info['min_replica'],
            '--max-replica', deployment_info['max_replica'],
            '--cpu-request', deployment_info['cpu_request'],
            '--cpu-limit', deployment_info['cpu_limit'],
            '--memory-request', deployment_info['memory_request'],
            ]
        )

    if result.exception:
        traceback.print_exception(*result.exc_info)

    test_deployed_model_version = result.output.split('\n')[0].split(' ')[-1]

    # Get latest deployed model's version
    merlin.set_url(deployment_info['url'], use_google_oauth=use_google_oauth)
    merlin.set_project(deployment_info['project'])
    merlin.set_model(model_name, ModelType.SKLEARN)

    merlin_active_model = merlin.active_model()
    all_versions = merlin_active_model.list_version()

    latest_version = all_versions[0]

    # Undeployment
    undeploy_result = runner.invoke(
        cli, [
            'undeploy',
            '--model-version', test_deployed_model_version,
            '--model-name', model_name,
            '--project', deployment_info['project'],
            '--url', deployment_info['url']
            ]
        )
    if result.exception:
        traceback.print_exception(*result.exc_info)

    planned_output = "Deleting deployment of model {} version {}".format(
        model_name,
        test_deployed_model_version
    )
    received_output = undeploy_result.output.split(' from')[0]

    assert latest_version._id == int(test_deployed_model_version)
    assert received_output == planned_output


@pytest.mark.cli
@pytest.mark.integration
def test_cli_scaffold_with_invalid_project(runner):
    result = runner.invoke(
        cli, [
            'scaffold',
            '--env', 'id',
            '--project', 'sample_project',
            '--model-name', 'pyfunc-prediction',
        ]
    )
    expected_output = '''Your project/model name contains invalid characters.\
                \nUse only the following characters\
                \n- Characters: a-z (Lowercase ONLY)\
                \n- Numbers: 0-9\
                \n- Symbols: -
            '''
    assert result.output.strip() == expected_output.strip()
    assert not os.path.exists('./pyfunc_prediction')


@pytest.mark.cli
@pytest.mark.integration
def test_cli_scaffold_with_invalid_model(runner):
    result = runner.invoke(
        cli, [
            'scaffold',
            '--env', 'id',
            '--project', 'project',
            '--model-name', 'pyfunc_prediction',
        ]
    )
    expected_output = '''Your project/model name contains invalid characters.\
                \nUse only the following characters\
                \n- Characters: a-z (Lowercase ONLY)\
                \n- Numbers: 0-9\
                \n- Symbols: -
            '''
    assert result.output.strip() == expected_output.strip()
    assert not os.path.exists('./pyfunc_prediction')
