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

# coding: utf-8


from __future__ import absolute_import

import signal
import sys

import merlin.autoscaling
import merlin.deployment_mode
import merlin.fluent
import merlin.resource_request
from merlin.version import VERSION as __version__

# Merlin URL
set_url = merlin.fluent.set_url
get_url = merlin.fluent.get_url

# Project
list_project = merlin.fluent.list_project
set_project = merlin.fluent.set_project
active_project = merlin.fluent.active_project

# Environment
list_environment = merlin.fluent.list_environment
get_environment = merlin.fluent.get_environment
get_default_environment = merlin.fluent.get_default_environment

# Model
set_model = merlin.fluent.set_model
active_model = merlin.fluent.active_model

# Model Version
new_model_version = merlin.fluent.new_model_version  # type:ignore
log_param = merlin.fluent.log_param
log_metric = merlin.fluent.log_metric
set_tag = merlin.fluent.set_tag
delete_tag = merlin.fluent.delete_tag
get_param = merlin.fluent.get_param
get_metric = merlin.fluent.get_metric
get_tag = merlin.fluent.get_tag
list_tag = merlin.fluent.list_tag
download_artifact = merlin.fluent.download_artifact

# Log Model Version
log_artifact = merlin.fluent.log_artifact
log_pyfunc_model = merlin.fluent.log_pyfunc_model
log_pytorch_model = merlin.fluent.log_pytorch_model
log_model = merlin.fluent.log_model
log_custom_model = merlin.fluent.log_custom_model

# Build image
build_image = merlin.fluent.build_image

# Model deployment - deploy active model version
deploy = merlin.fluent.deploy
undeploy = merlin.fluent.undeploy

# Model serving
set_traffic = merlin.fluent.set_traffic  # deprecated
serve_traffic = merlin.fluent.serve_traffic
stop_serving_traffic = merlin.fluent.stop_serving_traffic

# Model endpoints
list_model_endpoints = merlin.fluent.list_model_endpoints

# Definitions
ResourceRequest = merlin.resource_request.ResourceRequest
DeploymentMode = merlin.deployment_mode.DeploymentMode
AutoscalingPolicy = merlin.autoscaling.AutoscalingPolicy
MetricsType = merlin.autoscaling.MetricsType

# Batch
create_prediction_job = merlin.fluent.create_prediction_job

# Run server locally
run_pyfunc_model = merlin.pyfunc.run_pyfunc_model

__all__ = [
    "set_url",
    "get_url",
    "list_project",
    "set_project",
    "active_project",
    "list_environment",
    "get_environment",
    "get_default_environment",
    "set_model",
    "active_model",
    "new_model_version",
    "log_param",
    "log_metric",
    "set_tag",
    "delete_tag",
    "log_artifact",
    "log_pyfunc_model",
    "log_pytorch_model",
    "log_model",
    "build_image",
    "deploy",
    "undeploy",
    "set_traffic",
    "serve_traffic",
    "ResourceRequest",
    "DeploymentMode",
    "AutoscalingPolicy",
    "MetricsType",
    "create_prediction_job",
    "run_pyfunc_model",
]


def sigterm_handler(_signo, _stack_frame):
    # Raises SystemExit(0):
    sys.exit(0)


signal.signal(signal.SIGTERM, sigterm_handler)
