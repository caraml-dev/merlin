#!/usr/bin/python
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

import argparse
import os

from merlinpyspark.config import load
from merlinpyspark.model import create_model_udf
from merlinpyspark.sink import create_sink
from merlinpyspark.source import create_source
from mlflow import pyfunc
from pyspark import SparkConf, SparkContext
from pyspark.sql import SparkSession

try:
    import pyspark
except ImportError:
    import findspark

    findspark.init()
    import pyspark

DEFAULT_PARALLELISM = 2

parser = argparse.ArgumentParser(description="Run a PySpark job")
parser.add_argument(
    "--job-name",
    type=str,
    required=False,
    dest="job_name",
    help="The name of PySpark job",
    default="merlin-prediction-job",
)
parser.add_argument(
    "--spec-path",
    type=str,
    required=False,
    dest="spec_path",
    help="Path to prediction job yaml file",
)
parser.add_argument(
    "--dry-run-model",
    type=str,
    required=False,
    dest="dry_run_path",
    help="Path to model for dry run",
    default=None,
)
parser.add_argument(
    "--local",
    dest="local",
    action="store_true",
    required=False,
    help="flag to run locally",
    default=False,
)

args, _ = parser.parse_known_args()
print(f"Called with arguments: {args}")


def local_spark_session():
    conf = SparkConf()
    conf.set(
        "spark.jars",
        "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector"
        "-hadoop2-2.0.1.jar",
    )
    conf.set(
        "spark.jars.packages",
        "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.13.1-beta",
    )

    conf.set(
        "spark.jars",
        "https://github.com/caraml-dev/caraml-store/releases/download/caraml-store-spark/odps/v0.0.1-test/custom-dialect.jar,"
        "https://github.com/aliyun/aliyun-odps-jdbc/releases/download/v3.8.2/odps-jdbc-3.8.2-jar-with-dependencies.jar,"
        "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar"
    )

    sc = SparkContext(conf=conf)
    sc.setLogLevel("INFO")
    sc._jsc.hadoopConfiguration().set(
        "fs.gs.impl", "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"
    )
    sc._jsc.hadoopConfiguration().set(
        "fs.AbstractFileSystem.gs.impl",
        "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS",
    )
    sc._jsc.hadoopConfiguration().set(
        "google.cloud.auth.service.account.enable", "true"
    )

    sa_path = os.environ.get("GOOGLE_APPLICATION_CREDENTIALS")
    if sa_path is None:
        print("You must set GOOGLE_APPLICATION_CREDENTIALS to run locally")

    sc._jsc.hadoopConfiguration().set(
        "google.cloud.auth.service.account.json.keyfile", sa_path
    )

    spark = SparkSession.builder.config(conf=sc.getConf()).getOrCreate()
    return spark


def start(spec_path, spark):
    print(f"loading prediction job spec from: {spec_path}")
    job_spec = load(spec_path)

    # target parallelism is determined by number of executor instances * number of core per executor
    # for local case spark.executor.instances and spark.executor.cores are None and will fallback to DEFAULT_PARALLELISM
    instances = spark.sparkContext.getConf().get("spark.executor.instances")
    cores = spark.sparkContext.getConf().get("spark.executor.cores")
    target_parallelism = DEFAULT_PARALLELISM  # to handle local test
    if instances is not None and cores is not None:
        target_parallelism = int(instances) * int(cores)

    print(f"target parallelism: {target_parallelism}")

    data_source = create_source(spark, job_spec.source())
    df = data_source.load()
    features = list(data_source.features())
    if features is None:
        features = df.columns

    current_partition = df.rdd.getNumPartitions()
    if target_parallelism > current_partition:
        # Repartition the dataframe to have same number of partition as target_parallelism for better executor utilization
        print(
            f"repartition data from {df.rdd.getNumPartitions()} to {target_parallelism}"
        )
        df = df.repartition(target_parallelism)
    elif current_partition > target_parallelism:
        target_partition = (
            current_partition
            - (current_partition % target_parallelism)
            + target_parallelism
        )
        print(f"repartition data from {current_partition} to {target_partition}")
        df = df.repartition(target_partition)

    model_udf = create_model_udf(spark, job_spec.model(), features)
    data_sink = create_sink(spark, job_spec.sink())

    df = df.withColumn(job_spec.sink().result_column(), model_udf(*features))

    data_sink.save(df)

    print(f"The prediction job completed successfully!")


def main():
    if args.dry_run_path is not None:
        pyfunc.load_model(args.dry_run_path)
        print("Dry run success")
        exit(0)

    if args.local:
        spark = local_spark_session()
    else:
        spark = SparkSession.builder.appName(args.job_name).getOrCreate()
        spark.conf.set("spark.sql.execution.arrow.pyspark.enabled", "true")

    print(f"Spark configuration: {spark.sparkContext.getConf().getAll()}")

    start(args.spec_path, spark)


if __name__ == "__main__":
    main()
