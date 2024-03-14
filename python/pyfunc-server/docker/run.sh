#!/bin/bash
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

sigterm_handler() {
  echo "Terminating.."
  kill -TERM "$server_pid"
  wait "$server_pid"
}

sigint_handler() {
  echo "Terminating.."
  kill -INT "$server_pid"
  wait "$server_pid"
}

trap sigterm_handler SIGTERM
trap sigint_handler SIGINT

source activate merlin-model
merlin-pyfunc-server --model_dir ./model &

server_pid=$!
wait "$server_pid"
