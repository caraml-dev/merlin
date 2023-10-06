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
import asyncio
import logging
import traceback

import uvloop

from pyfuncserver.config import Config
from pyfuncserver.model.model import PyFuncModel
from pyfuncserver.server import PyFuncServer
from pyfuncserver.utils.contants import ERR_DRY_RUN

parser = argparse.ArgumentParser()
parser.add_argument('--model_dir', required=True,
                    help='A URI pointer to the model binary')
parser.add_argument('--dry_run', default=False, action='store_true', required=False,
                    help="Dry run pyfunc server by loading the specified model "
                         "in --model_dir without starting webserver")
args, _ = parser.parse_known_args()

logging.getLogger('tornado.access').disabled = True

import os
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.trace.export import ConsoleSpanExporter
from opentelemetry.exporter.jaeger.thrift import JaegerExporter
from opentelemetry.sdk.resources import SERVICE_NAME, Resource

def setup_tracer():
    if os.getenv("JAEGER_DISABLED", "true") == "false":
        provider = TracerProvider()
        processor = BatchSpanProcessor(ConsoleSpanExporter())
        provider.add_span_processor(processor)

        trace.set_tracer_provider(
            TracerProvider(resource=Resource.create({SERVICE_NAME: os.getenv("K_REVISION", "pyfunc-service")}))
        )
        jaeger_exporter = JaegerExporter(
            collector_endpoint=os.getenv("JAEGER_COLLECTOR_URL", "http://localhost:14268/api/traces")
        )
        trace.get_tracer_provider().add_span_processor(
            BatchSpanProcessor(jaeger_exporter)
        )

if __name__ == "__main__":
    setup_tracer()
    
    # use uvloop as the event loop
    asyncio.set_event_loop_policy(uvloop.EventLoopPolicy())

    config = Config(args.model_dir)
    logging.basicConfig(level=config.log_level)
    # load model
    model = PyFuncModel(config.model_manifest)

    try:
        model.load()
    except Exception as e:
        logging.error(f"Unable to initalize model")
        logging.error(traceback.format_exc())
        logging.error(ERR_DRY_RUN)
        exit(1)

    if args.dry_run:
        logging.info("dry run success")
        exit(0)

    PyFuncServer(config).start(model)
