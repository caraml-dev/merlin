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

import json
import logging
from http import HTTPStatus
from typing import Any, Dict

import orjson
import tornado.web

from pyfuncserver.model.model import PyFuncModel
from merlin.pyfunc import PyFuncOutput


class PredictHandler(tornado.web.RequestHandler):
    def initialize(self, models):
        self.models = models  # pylint:disable=attribute-defined-outside-init
        self.publisher = self.application.publisher

    def get_model(self, full_name: str):
        if full_name not in self.models:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.NOT_FOUND,
                reason="Model with full name %s does not exist." % full_name
            )
        model = self.models[full_name]
        if not model.ready:
            model.load()
        return model

    def get_headers(self, request):
        return {k: v for (k, v) in request.headers.get_all()}

    def validate(self, request):
        try:
            body = orjson.loads(request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e
            )
        return body

    def post(self, full_name: str):
        model = self.get_model(full_name)

        request = self.validate(self.request)
        headers = self.get_headers(self.request)

        output = model.predict(request, headers=headers)
        response_json = output
        output_is_pyfunc_output = isinstance(response_json, PyFuncOutput)
        if output_is_pyfunc_output:
            response_json = output.http_response
         
        response_json = orjson.dumps(response_json)
        self.write(response_json)
        self.set_header("Content-Type", "application/json; charset=UTF-8")

        if self.publisher is not None and output_is_pyfunc_output and output.contains_prediction_log():
            tornado.ioloop.IOLoop.current().spawn_callback(self.publisher.publish, output)

    def write_error(self, status_code: int, **kwargs: Any) -> None:
        logging.error(self._reason)
        self.write({"status_code": status_code, "reason": self._reason})


class LivenessHandler(tornado.web.RequestHandler):  # pylint:disable=too-few-public-methods
    def get(self):
        self.write("Alive")


class HealthHandler(tornado.web.RequestHandler):
    def initialize(self, models: Dict[str, PyFuncModel]):
        self.models = models  # pylint:disable=attribute-defined-outside-init

    def get(self, full_name: str):
        if full_name not in self.models:
            raise tornado.web.HTTPError(
                status_code=404,
                reason="Model with full name %s does not exist." % full_name
            )

        model = self.models[full_name]
        self.write(json.dumps({
            "name": model.full_name,
            "ready": model.ready
        }))
