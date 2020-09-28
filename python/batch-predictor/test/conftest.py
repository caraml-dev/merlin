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
import shutil

import pytest
from google.cloud import bigquery
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession

shutil.rmtree("mlruns", ignore_errors=True)

@pytest.fixture(scope="session")
def spark_session(request):
    conf = SparkConf()
    conf.set("spark.jars",
             "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector"
             "-hadoop2-2.0.1.jar")
    conf.set("spark.jars.packages",
             "com.google.cloud.spark:spark-bigquery-with-dependencies_2.11:0.13.1-beta")
    sc = SparkContext(conf=conf)

    sc._jsc.hadoopConfiguration().set("fs.gs.impl",
                                      "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem")
    sc._jsc.hadoopConfiguration().set("fs.AbstractFileSystem.gs.impl",
                                      "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS")
    sc._jsc.hadoopConfiguration().set(
        "google.cloud.auth.service.account.enable", "true")

    sa_path = os.environ.get("GOOGLE_APPLICATION_CREDENTIALS")
    if sa_path is not None:
        sc._jsc.hadoopConfiguration().set(
            "google.cloud.auth.service.account.json.keyfile",
            sa_path)

    spark = SparkSession.builder \
        .config(conf=sc.getConf()) \
        .getOrCreate()

    request.addfinalizer(lambda: spark.stop())

    return spark


@pytest.fixture
def bq():
    return bigquery.Client()
