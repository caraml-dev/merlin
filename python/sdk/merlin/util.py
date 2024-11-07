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

import re
import os
import boto3
from urllib.parse import urlparse
from google.cloud import storage
from os.path import dirname
from os import makedirs
from typing import Optional, Any

def guess_mlp_ui_url(mlp_api_url: str) -> str:
    raw_url = mlp_api_url.replace("/api", "")
    return get_url(raw_url)


def get_url(raw_url, scheme='http'):
    """
    Get url or prefix with default scheme if it doesn't have one
    """
    parsed_url = urlparse(raw_url)
    if not parsed_url.scheme:
        # if scheme is not provided then assume it's http
        raw_url = f"{scheme}://{raw_url}"
    return raw_url


def autostr(cls):
    def __str__(self):
        return '%s(%s)' % (
            type(self).__name__,
            ', '.join('%s=%s' % item for item in vars(self).items())
        )

    cls.__str__ = __str__
    cls.__repr__ = __str__
    return cls


def valid_name_check(input_name: str) -> bool:
    """
    Checks if inputted name for project and model is url-friendly
        - has to be lower case
        - can only contain character (a-z) number (0-9) and some limited symbols
    """
    # allowed characters to be included in pattern after backslash
    pattern = r'[-a-z0-9]+'

    matching_group = None
    if re.search(pattern, input_name):
        match = re.search(pattern, input_name)
        if match is None:
            return False
        matching_group = match.group(0)
    return matching_group == input_name


def get_blob_storage_scheme(artifact_uri: str) -> str:
    parsed_result = urlparse(artifact_uri)
    return parsed_result.scheme


def get_bucket_name(artifact_uri: str) -> str:
    parsed_result = urlparse(artifact_uri)
    return parsed_result.netloc


def get_artifact_path(artifact_uri: str) -> str:
    parsed_result = urlparse(artifact_uri)
    return parsed_result.path.strip("/")


def download_files_from_blob_storage(artifact_uri: str, destination_path: str):
    makedirs(destination_path, exist_ok=True)

    storage_schema = get_blob_storage_scheme(artifact_uri)
    bucket_name = get_bucket_name(artifact_uri)
    path = get_artifact_path(artifact_uri)

    if storage_schema == "gs":
        client = storage.Client()
        bucket = client.get_bucket(bucket_name)
        blobs = bucket.list_blobs(prefix=path)
        for blob in blobs:
            # Get only the path after .../artifacts/model
            # E.g.
            # Some blob looks like this mlflow/3/ad8f15a4023f461796955f71e1152bac/artifacts/model/1/saved_model.pb
            # we only want to extract 1/saved_model.pb
            artifact_path = os.path.join(*blob.name.split("/")[5:])
            dir = os.path.join(destination_path, dirname(artifact_path))
            makedirs(dir, exist_ok=True)
            blob.download_to_filename(os.path.join(destination_path, artifact_path))
    elif storage_schema == "s3":
        client = boto3.client("s3")
        bucket = client.list_objects_v2(Prefix=path, Bucket=bucket_name)["Contents"]
        for s3_object in bucket:
            # we do this because the list_objects_v2 method lists all subdirectories in addition to files
            if not s3_object['Key'].endswith('/'):
                # Get only the path after .../artifacts/model
                # E.g.
                # Some blob looks like this mlflow/3/ad8f15a4023f461796955f71e1152bac/artifacts/model/1/saved_model.pb
                # we only want to extract 1/saved_model.pb
                object_paths = s3_object['Key'].split("/")[5:]
                if len(object_paths) != 0:
                    artifact_path = os.path.join(*object_paths)
                    os.makedirs(os.path.join(destination_path, dirname(artifact_path)), exist_ok=True)
                    client.download_file(bucket_name, s3_object['Key'], os.path.join(destination_path, artifact_path))


def extract_optional_value_with_default(opt: Optional[Any], default: Any) -> Any:
    if opt is not None:
        return opt
    return default