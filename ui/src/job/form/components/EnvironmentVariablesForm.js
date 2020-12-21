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

import React, { useEffect, useState, useCallback, Fragment } from "react";

import {
  EuiButtonIcon,
  EuiFieldText,
  EuiInMemoryTable,
  EuiTitle
} from "@elastic/eui";
import PropTypes from "prop-types";

require("../../../assets/scss/EnvironmentVariables.scss");

export const EnvironmentVariablesForm = ({ variables, onChange }) => {
  const [items, setItems] = useState([
    ...variables.map((v, idx) => ({ idx, ...v })),
    { idx: variables.length }
  ]);

  const setVars = useCallback(items => onChange(items), [onChange]);

  useEffect(() => {
    const trimmedVars = [
      ...items
        .slice(0, items.length - 1)
        .map(item => ({ name: item.name.trim(), value: item.value }))
    ];
    if (JSON.stringify(variables) !== JSON.stringify(trimmedVars)) {
      setVars(trimmedVars);
    }
  }, [variables, items, setVars]);

  const removeRow = idx => {
    items.splice(idx, 1);
    setItems([...items.map((v, idx) => ({ ...v, idx }))]);
  };

  const onChangeRow = (idx, field) => {
    return e => {
      items[idx] = { ...items[idx], [field]: e.target.value };

      setItems(_ =>
        field === "name" &&
        items[items.length - 1].name &&
        items[items.length - 1].name.trim()
          ? [...items, { idx: items.length }]
          : [...items]
      );
    };
  };

  const columns = [
    {
      name: "Name",
      field: "name",
      width: "45%",
      render: (name, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Name"
          value={name || ""}
          onChange={onChangeRow(item.idx, "name")}
        />
      )
    },
    {
      name: "Value",
      field: "value",
      width: "45%",
      render: (value, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Value"
          value={value || ""}
          onChange={onChangeRow(item.idx, "value")}
        />
      )
    },
    {
      width: "10%",
      actions: [
        {
          render: item => {
            return item.idx < items.length - 1 ? (
              <EuiButtonIcon
                size="s"
                color="danger"
                iconType="trash"
                onClick={() => removeRow(item.idx)}
                aria-label="Remove variable"
              />
            ) : (
              <div />
            );
          }
        }
      ]
    }
  ];

  return (
    <Fragment>
      <EuiTitle size="xs">
        <h4>Environment Variables</h4>
      </EuiTitle>

      <EuiInMemoryTable
        className="EnvVariables"
        columns={columns}
        items={items}
        hasActions={true}
      />
    </Fragment>
  );
};

EnvironmentVariablesForm.propTypes = {
  variables: PropTypes.array,
  onChange: PropTypes.func
};
