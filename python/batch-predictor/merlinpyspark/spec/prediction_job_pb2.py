# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: spec/prediction_job.proto
# Protobuf Python Version: 4.25.6
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x19spec/prediction_job.proto\x12\x11merlin.batch.spec\"\xd3\x03\n\rPredictionJob\x12\x0f\n\x07version\x18\x01 \x01(\t\x12\x0c\n\x04kind\x18\x02 \x01(\t\x12\x0c\n\x04name\x18\x03 \x01(\t\x12<\n\x0f\x62igquery_source\x18\x0b \x01(\x0b\x32!.merlin.batch.spec.BigQuerySourceH\x00\x12\x32\n\ngcs_source\x18\x0c \x01(\x0b\x32\x1c.merlin.batch.spec.GcsSourceH\x00\x12@\n\x11maxcompute_source\x18\r \x01(\x0b\x32#.merlin.batch.spec.MaxComputeSourceH\x00\x12\'\n\x05model\x18\x15 \x01(\x0b\x32\x18.merlin.batch.spec.Model\x12\x38\n\rbigquery_sink\x18\x1f \x01(\x0b\x32\x1f.merlin.batch.spec.BigQuerySinkH\x01\x12.\n\x08gcs_sink\x18  \x01(\x0b\x32\x1a.merlin.batch.spec.GcsSinkH\x01\x12<\n\x0fmaxcompute_sink\x18! \x01(\x0b\x32!.merlin.batch.spec.MaxComputeSinkH\x01\x42\x08\n\x06sourceB\x06\n\x04sink\"\xa2\x01\n\x0e\x42igQuerySource\x12\r\n\x05table\x18\x01 \x01(\t\x12\x10\n\x08\x66\x65\x61tures\x18\x02 \x03(\t\x12?\n\x07options\x18\x03 \x03(\x0b\x32..merlin.batch.spec.BigQuerySource.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\xc5\x01\n\tGcsSource\x12-\n\x06\x66ormat\x18\x01 \x01(\x0e\x32\x1d.merlin.batch.spec.FileFormat\x12\x0b\n\x03uri\x18\x02 \x01(\t\x12\x10\n\x08\x66\x65\x61tures\x18\x03 \x03(\t\x12:\n\x07options\x18\x04 \x03(\x0b\x32).merlin.batch.spec.GcsSource.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\xb8\x01\n\x10MaxComputeSource\x12\r\n\x05table\x18\x01 \x01(\t\x12\x10\n\x08\x65ndpoint\x18\x02 \x01(\t\x12\x10\n\x08\x66\x65\x61tures\x18\x03 \x03(\t\x12\x41\n\x07options\x18\x04 \x03(\x0b\x32\x30.merlin.batch.spec.MaxComputeSource.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\xcc\x02\n\x05Model\x12*\n\x04type\x18\x01 \x01(\x0e\x32\x1c.merlin.batch.spec.ModelType\x12\x0b\n\x03uri\x18\x02 \x01(\t\x12\x34\n\x06result\x18\x03 \x01(\x0b\x32$.merlin.batch.spec.Model.ModelResult\x12\x36\n\x07options\x18\x04 \x03(\x0b\x32%.merlin.batch.spec.Model.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x1al\n\x0bModelResult\x12+\n\x04type\x18\x01 \x01(\x0e\x32\x1d.merlin.batch.spec.ResultType\x12\x30\n\titem_type\x18\x02 \x01(\x0e\x32\x1d.merlin.batch.spec.ResultType\"\xeb\x01\n\x0c\x42igQuerySink\x12\r\n\x05table\x18\x01 \x01(\t\x12\x16\n\x0estaging_bucket\x18\x02 \x01(\t\x12\x15\n\rresult_column\x18\x03 \x01(\t\x12.\n\tsave_mode\x18\x04 \x01(\x0e\x32\x1b.merlin.batch.spec.SaveMode\x12=\n\x07options\x18\x05 \x03(\x0b\x32,.merlin.batch.spec.BigQuerySink.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\xf6\x01\n\x07GcsSink\x12-\n\x06\x66ormat\x18\x01 \x01(\x0e\x32\x1d.merlin.batch.spec.FileFormat\x12\x0b\n\x03uri\x18\x02 \x01(\t\x12\x15\n\rresult_column\x18\x03 \x01(\t\x12.\n\tsave_mode\x18\x04 \x01(\x0e\x32\x1b.merlin.batch.spec.SaveMode\x12\x38\n\x07options\x18\x05 \x03(\x0b\x32\'.merlin.batch.spec.GcsSink.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\"\xe9\x01\n\x0eMaxComputeSink\x12\r\n\x05table\x18\x01 \x01(\t\x12\x10\n\x08\x65ndpoint\x18\x02 \x01(\t\x12\x15\n\rresult_column\x18\x03 \x01(\t\x12.\n\tsave_mode\x18\x04 \x01(\x0e\x32\x1b.merlin.batch.spec.SaveMode\x12?\n\x07options\x18\x05 \x03(\x0b\x32..merlin.batch.spec.MaxComputeSink.OptionsEntry\x1a.\n\x0cOptionsEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01*Q\n\nResultType\x12\n\n\x06\x44OUBLE\x10\x00\x12\t\n\x05\x46LOAT\x10\x01\x12\x0b\n\x07INTEGER\x10\x02\x12\x08\n\x04LONG\x10\x03\x12\n\n\x06STRING\x10\x04\x12\t\n\x05\x41RRAY\x10\n*\x7f\n\tModelType\x12\x16\n\x12INVALID_MODEL_TYPE\x10\x00\x12\x0b\n\x07XGBOOST\x10\x01\x12\x0e\n\nTENSORFLOW\x10\x02\x12\x0b\n\x07SKLEARN\x10\x03\x12\x0b\n\x07PYTORCH\x10\x04\x12\x08\n\x04ONNX\x10\x05\x12\n\n\x06PYFUNC\x10\x06\x12\r\n\tPYFUNC_V2\x10\x07*O\n\nFileFormat\x12\x17\n\x13INVALID_FILE_FORMAT\x10\x00\x12\x07\n\x03\x43SV\x10\x01\x12\x0b\n\x07PARQUET\x10\x02\x12\x08\n\x04\x41VRO\x10\x03\x12\x08\n\x04JSON\x10\x04*O\n\x08SaveMode\x12\x11\n\rERRORIFEXISTS\x10\x00\x12\r\n\tOVERWRITE\x10\x01\x12\n\n\x06\x41PPEND\x10\x02\x12\n\n\x06IGNORE\x10\x03\x12\t\n\x05\x45RROR\x10\x04\x42\x66\n\x1b\x63om.gojek.merlin.batch.specB\x12PredictionJobProtoP\x01Z1github.com/caraml-dev/merlin-pyspark-app/pkg/specb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'spec.prediction_job_pb2', _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
  _globals['DESCRIPTOR']._options = None
  _globals['DESCRIPTOR']._serialized_options = b'\n\033com.gojek.merlin.batch.specB\022PredictionJobProtoP\001Z1github.com/caraml-dev/merlin-pyspark-app/pkg/spec'
  _globals['_BIGQUERYSOURCE_OPTIONSENTRY']._options = None
  _globals['_BIGQUERYSOURCE_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_GCSSOURCE_OPTIONSENTRY']._options = None
  _globals['_GCSSOURCE_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_MAXCOMPUTESOURCE_OPTIONSENTRY']._options = None
  _globals['_MAXCOMPUTESOURCE_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_MODEL_OPTIONSENTRY']._options = None
  _globals['_MODEL_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_BIGQUERYSINK_OPTIONSENTRY']._options = None
  _globals['_BIGQUERYSINK_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_GCSSINK_OPTIONSENTRY']._options = None
  _globals['_GCSSINK_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_MAXCOMPUTESINK_OPTIONSENTRY']._options = None
  _globals['_MAXCOMPUTESINK_OPTIONSENTRY']._serialized_options = b'8\001'
  _globals['_RESULTTYPE']._serialized_start=2128
  _globals['_RESULTTYPE']._serialized_end=2209
  _globals['_MODELTYPE']._serialized_start=2211
  _globals['_MODELTYPE']._serialized_end=2338
  _globals['_FILEFORMAT']._serialized_start=2340
  _globals['_FILEFORMAT']._serialized_end=2419
  _globals['_SAVEMODE']._serialized_start=2421
  _globals['_SAVEMODE']._serialized_end=2500
  _globals['_PREDICTIONJOB']._serialized_start=49
  _globals['_PREDICTIONJOB']._serialized_end=516
  _globals['_BIGQUERYSOURCE']._serialized_start=519
  _globals['_BIGQUERYSOURCE']._serialized_end=681
  _globals['_BIGQUERYSOURCE_OPTIONSENTRY']._serialized_start=635
  _globals['_BIGQUERYSOURCE_OPTIONSENTRY']._serialized_end=681
  _globals['_GCSSOURCE']._serialized_start=684
  _globals['_GCSSOURCE']._serialized_end=881
  _globals['_GCSSOURCE_OPTIONSENTRY']._serialized_start=635
  _globals['_GCSSOURCE_OPTIONSENTRY']._serialized_end=681
  _globals['_MAXCOMPUTESOURCE']._serialized_start=884
  _globals['_MAXCOMPUTESOURCE']._serialized_end=1068
  _globals['_MAXCOMPUTESOURCE_OPTIONSENTRY']._serialized_start=635
  _globals['_MAXCOMPUTESOURCE_OPTIONSENTRY']._serialized_end=681
  _globals['_MODEL']._serialized_start=1071
  _globals['_MODEL']._serialized_end=1403
  _globals['_MODEL_OPTIONSENTRY']._serialized_start=635
  _globals['_MODEL_OPTIONSENTRY']._serialized_end=681
  _globals['_MODEL_MODELRESULT']._serialized_start=1295
  _globals['_MODEL_MODELRESULT']._serialized_end=1403
  _globals['_BIGQUERYSINK']._serialized_start=1406
  _globals['_BIGQUERYSINK']._serialized_end=1641
  _globals['_BIGQUERYSINK_OPTIONSENTRY']._serialized_start=635
  _globals['_BIGQUERYSINK_OPTIONSENTRY']._serialized_end=681
  _globals['_GCSSINK']._serialized_start=1644
  _globals['_GCSSINK']._serialized_end=1890
  _globals['_GCSSINK_OPTIONSENTRY']._serialized_start=635
  _globals['_GCSSINK_OPTIONSENTRY']._serialized_end=681
  _globals['_MAXCOMPUTESINK']._serialized_start=1893
  _globals['_MAXCOMPUTESINK']._serialized_end=2126
  _globals['_MAXCOMPUTESINK_OPTIONSENTRY']._serialized_start=635
  _globals['_MAXCOMPUTESINK_OPTIONSENTRY']._serialized_end=681
# @@protoc_insertion_point(module_scope)
