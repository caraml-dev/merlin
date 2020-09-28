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

import React, { useContext } from "react";
import { AuthContext } from "@gojek/mlp-ui";
import { LazyLog } from "react-lazylog";

const StreamLog = ({ logUrl }) => {
  const authCtx = useContext(AuthContext);
  const fetchOptions = {
    headers: {
      Authorization: `Bearer ${authCtx.state.accessToken}`
    }
  };

  return (
    <LazyLog
      caseInsensitive
      enableSearch
      fetchOptions={fetchOptions}
      extraLines={1}
      height={640}
      url={logUrl}
      selectableLines
      follow
      stream
    />
  );
};

export default StreamLog;
