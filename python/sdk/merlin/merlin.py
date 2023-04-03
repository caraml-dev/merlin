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

import click
import warnings
import os
import merlin
from merlin.model import ModelType
from cookiecutter.main import cookiecutter
from merlin.util import valid_name_check

warnings.filterwarnings('ignore')

@click.group()
def cli():
    """
    A simple command line tool.

    The Merlin CLI assumes that you already have a serialized model.\n
    To see the options for each command: merlin COMMAND --help
    """
    pass

@cli.command('deploy', short_help='Deploy the model')
@click.option('--env', '-e', required=True, help='The environment of model deployment')
@click.option('--url', '-u', required=True, help='The endpoint of model deployment')
@click.option('--project', '-p', required=True, help='The project name of model deployment')
@click.option('--model-dir', '-m', required=True, help='The directory with model for deployment')
@click.option('--model-name', '-n', required=True, help='The model name for deployment')
@click.option('--model-type', '-t', required=True, help='The type of machine learning algorithm')
@click.option('--min-replica', required=False, help='The minimum number of replicas to create for this deployment')
@click.option('--max-replica', required=False, help='The maximum number of replicas to create for this deployment')
@click.option('--cpu-request', required=False, help='The CPU resource requirement requests for this deployment. Example: 100m.')
@click.option('--memory-request', required=False, help='The memory resource requirement requests for this deployment. Example: 256Mi.')
def deploy(env, model_name, model_type, model_dir, project, url,
           min_replica, max_replica, cpu_request, memory_request):

    merlin.set_url(url)

    target_env = merlin.get_environment(env)

    resource_request = target_env.default_resource_request
    if min_replica is not None:
        resource_request.min_replica = int(min_replica)
    if max_replica is not None:
        resource_request.max_replica = int(max_replica)
    if cpu_request is not None:
        resource_request.cpu_request = cpu_request
    if memory_request is not None:
        resource_request.memory_request = memory_request

    merlin.set_project(project)
    merlin.set_model(model_name, ModelType(model_type))

    with merlin.new_model_version() as v:
        merlin.log_model(model_dir=model_dir)

    try:
        endpoint = merlin.deploy(v, env, resource_request)
        if endpoint:
            print('Model deployed to {}'.format(endpoint))
    except Exception as e:
        print(e)

@cli.command('undeploy', short_help='Undeploy the model')
@click.option('--url', '-u', required=True, help='The endpoint of model deployment')
@click.option('--project', '-p', required=True, help='The project name of model deployment')
@click.option('--model-name', '-n', required=True, help='The model name for deployment')
@click.option('--model-version', '-v', required=True, help='The model version for deployment')
def undeploy(model_name, model_version, project, url):

    merlin.set_url(url)
    merlin.set_project(project)
    merlin.set_model(model_name)

    merlin_active_model = merlin.active_model()
    all_versions = merlin_active_model.list_version()

    try:
        wanted_model_info = [model_info for model_info in all_versions if model_info._id == int(model_version)][0]
    except Exception as e:
        print(e)
        print('Model Version {} is not found.'.format(model_version))

    try:
        merlin.undeploy(wanted_model_info)
    except Exception as e:
        print(e)


@cli.command('scaffold', short_help='Generate PyFunc project')
@click.option('--project', '-p', required=True, help='The merlin project name of PyFunc server')
@click.option('--model-name', '-m', required=True, help='The model name which will be listed in merlin')
@click.option('--env', '-e', required=True, help='The environment which PyFunc server will be deployed, available environment are id and global')
def scaffold(project, model_name, env):
    if not valid_name_check(project) or not valid_name_check(model_name):
        print(
            '''Your project/model name contains invalid characters.\
                \nUse only the following characters\
                \n- Characters: a-z (Lowercase ONLY)\
                \n- Numbers: 0-9\
                \n- Symbols: -
            '''
        )
    else:
        try:
            cookiecutter("git@github.com:caraml-dev/merlin/python/pyfunc-scaffolding",
                checkout="tags/v0.1",
                no_input=True, directory="python/pyfunc-scaffolding",
                extra_context={'project_name': project, 'model_name': model_name, 'environment_name': env})

        except Exception as e:
            print(e)


if __name__ == "__main__":
    cli()
