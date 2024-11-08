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

from merlin.util import guess_mlp_ui_url, valid_name_check, get_blob_storage_scheme, get_bucket_name, get_artifact_path
import pytest


@pytest.mark.unit
def test_name_check():
    invalid_names = [
        'test-name,', 'test,name', 'test.name'
    ]

    valid_names = [
        'test-name', 'testname'
    ]

    for name in invalid_names:
        result = valid_name_check(name)
        assert result is False
    for name in valid_names:
        result = valid_name_check(name)
        assert result is True


@pytest.mark.unit
def test_get_blob_storage_scheme():
    gcs_artifact_uri = 'gs://some-bucket/mlflow/81/ddd'
    assert get_blob_storage_scheme(gcs_artifact_uri) == 'gs'

    s3_artifact_uri = 's3://some-bucket/mlflow/81/ddd'
    assert get_blob_storage_scheme(s3_artifact_uri) == 's3'


@pytest.mark.unit
def test_get_bucket_name():
    artifact_uri = 'gs://some-bucket/mlflow/81/ddd'
    assert get_bucket_name(artifact_uri) == 'some-bucket'


@pytest.mark.unit
def test_get_artifact_path():
    artifact_uri = 'gs://some-bucket/mlflow/81/ddd'
    assert get_artifact_path(artifact_uri) == 'mlflow/81/ddd'

    double_slash_uri_path = 'gs://some-bucket//mlflow/81/ddd'
    assert get_artifact_path(double_slash_uri_path) == 'mlflow/81/ddd'


@pytest.mark.parametrize("mlp_api_url,expected", [
    ("http://console.mydomain.com/merlin/api", "http://console.mydomain.com/merlin"),
    ("https://console.mydomain.com/merlin/api", "https://console.mydomain.com/merlin"),
    ("console.mydomain.com/merlin/api", "http://console.mydomain.com/merlin"),
])
@pytest.mark.unit
def test_guess_mlp_ui_url(mlp_api_url, expected):
    assert guess_mlp_ui_url(mlp_api_url) == expected
