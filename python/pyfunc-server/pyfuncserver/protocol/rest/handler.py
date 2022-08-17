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
from kfserving.kfmodel import KFModel


class HTTPHandler(tornado.web.RequestHandler):
    def initialize(self, models: Dict[str, KFModel]):
        self.models = models  # pylint:disable=attribute-defined-outside-init

    def get_model(self, name: str):
        if name not in self.models:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.NOT_FOUND,
                reason="Model with name %s does not exist." % name
            )
        model = self.models[name]
        if not model.ready:
            model.load()
        return model

    def get_headers(self, request):
        return {k:v for (k, v) in request.headers.get_all()}

    def validate(self, request):
        try:
            body = orjson.loads(request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e
            )
        return body


class CustomPredictHandler(HTTPHandler):
    def post(self, name: str):
        model = self.get_model(name)

        request = self.validate(self.request)
        headers = self.get_headers(self.request)

        request = model.preprocess(request)
        response = model.predict(request, headers=headers)
        response = model.postprocess(response)

        response_json = orjson.dumps(response)
        self.write(response_json)
        self.set_header("Content-Type", "application/json; charset=UTF-8")

    def write_error(self, status_code: int, **kwargs: Any) -> None:
        logging.error(self._reason)
        self.write({"status_code": status_code, "reason": self._reason})
