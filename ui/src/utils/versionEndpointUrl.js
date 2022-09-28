/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import PropTypes from "prop-types";

export const versionEndpointUrl = (url, protocol) => {
  if (protocol && protocol === "HTTP_JSON") {
    return url.includes(":predict") ? url : url + ":predict";
  }

  // UPI_V1
  return url;
};

versionEndpointUrl.propTypes = {
  url: PropTypes.string.isRequired,
};
