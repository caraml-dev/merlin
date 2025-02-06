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

import numpy as np
import pandas
from mlflow.tracking.artifact_utils import _download_artifact_from_uri
from mlflow.utils.file_utils import TempDir
from pyspark.sql import SparkSession

from merlinpyspark.config import ModelConfig
from merlinpyspark.spec.prediction_job_pb2 import ModelType


def create_model_udf(spark_session: SparkSession, model_config: ModelConfig, features):
    if model_config.model_type() == ModelType.Name(ModelType.PYFUNC_V2):
        return spark_udf(spark_session, model_config.model_uri(), features,
                         result_type=model_config.result_type())

    raise ValueError(f"model type not supported: {model_config.model_type()}")


def spark_udf(spark, model_uri, features, result_type="double"):
    """
    Create spark pandas udf given the model uri
    :param spark: spark context
    :param model_uri: path to model
    :param features: list containing the feature names
    :param result_type: result type of the model
    :return:
    """
    # Scope Spark import to this method so users don't need pyspark to use non-Spark-related
    # functionality.
    from mlflow.pyfunc.spark_model_cache import SparkModelCache
    from pyspark.sql.functions import pandas_udf
    from pyspark.sql.types import _parse_datatype_string
    from pyspark.sql.types import ArrayType, DataType
    from pyspark.sql.types import DoubleType, IntegerType, FloatType, LongType, StringType

    if not isinstance(result_type, DataType):
        result_type = _parse_datatype_string(result_type)

    elem_type = result_type
    if isinstance(elem_type, ArrayType):
        elem_type = elem_type.elementType

    supported_types = [IntegerType, LongType, FloatType, DoubleType, StringType]

    if not any([isinstance(elem_type, x) for x in supported_types]):
        raise ValueError(
            "Invalid result_type '{}'. Result type can only be one of or an array of one "
            "of the following types types: {}".format(str(elem_type), str(supported_types)))

    with TempDir() as local_tmpdir:
        local_model_path = _download_artifact_from_uri(
            artifact_uri=model_uri, output_path=local_tmpdir.path())
        archive_path = SparkModelCache.add_local_model(spark, local_model_path)

    def predict(*args):
        model, _ = SparkModelCache.get_or_load(archive_path)
        schema = {features[i]: arg for i, arg in enumerate(args)}
        pdf = None
        for x in args:
            if type(x) == pandas.DataFrame:
                if len(args) != 1:
                    raise Exception("If passing a StructType column, there should be only one "
                                    "input column, but got %d" % len(args))
                pdf = x
        if pdf is None:
            pdf = pandas.DataFrame(schema)
        result = model.predict(pdf)
        if not isinstance(result, pandas.DataFrame):
            result = pandas.DataFrame(data=result)

        elif type(elem_type) == IntegerType:
            result = result.select_dtypes([np.byte, np.ubyte, np.short, np.ushort,
                                           np.int32]).astype(np.int32)

        elif type(elem_type) == LongType:
            result = result.select_dtypes([np.byte, np.ubyte, np.short, np.ushort, np.int, np.long])

        elif type(elem_type) == FloatType:
            result = result.select_dtypes(include=(np.number,)).astype(np.float32)

        elif type(elem_type) == DoubleType:
            result = result.select_dtypes(include=(np.number,)).astype(np.float64)

        if len(result.columns) == 0:
            raise ValueError(
                "The the model did not produce any values compatible with the requested "
                "type '{}'. Consider requesting udf with StringType or "
                "Arraytype(StringType).".format(str(elem_type)))

        if type(elem_type) == StringType:
            result = result.applymap(str)

        if type(result_type) == ArrayType:
            return pandas.Series([row[1].values for row in result.iterrows()])
        else:
            return result[result.columns[0]]

    return pandas_udf(predict, result_type)
