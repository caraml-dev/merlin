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

import React from "react";
import PropTypes from "prop-types";
import { EuiInMemoryTable, EuiText } from "@elastic/eui";

export const EnvVarsConfigTable = ({ variables }) => {
  const columns = [
    {
      field: "name",
      name: "Name",
      width: "20%",
      sortable: true
    },
    {
      field: "value",
      name: "Value",
      width: "80%",
      sortable: true
    }
  ];

  const sorting = {
    enableAllColumns: true
  };

  return variables.length ? (
    <EuiInMemoryTable
      items={variables}
      columns={columns}
      itemId="name"
      sorting={sorting}
    />
  ) : (
    <EuiText size="s" color="subdued">
      Not available
    </EuiText>
  );
};

EnvVarsConfigTable.propTypes = {
  variables: PropTypes.array.isRequired
};
