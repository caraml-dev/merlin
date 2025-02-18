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

import React, { Fragment, useCallback, useEffect, useState } from "react";

import {
  EuiButtonIcon,
  EuiFieldText,
  EuiInMemoryTable,
  EuiTitle,
  EuiSuperSelect
} from "@elastic/eui";
import PropTypes from "prop-types";
import { useParams  } from "react-router-dom";
import { useMerlinApi } from "../../../../hooks/useMerlinApi";

require("../../../../assets/scss/Secrets.scss");

export const SecretsForm = ({ variables, onChange }) => {
  const [items, setItems] = useState([
    ...variables.map((v, idx) => ({ idx, ...v })),
    { idx: variables.length },
  ]);

  const setVars = useCallback((items) => onChange(items), [onChange]);

  useEffect(() => {
    const trimmedVars = [
      ...items
        .slice(0, items.length - 1)
        .map((item) => ({ mlp_secret_name: item.mlp_secret_name.trim(), env_var_name: item.env_var_name })),
    ];
    if (JSON.stringify(variables) !== JSON.stringify(trimmedVars)) {
      setVars(trimmedVars);
    }
  }, [variables, items, setVars]);

  const removeRow = (idx) => {
    items.splice(idx, 1);
    setItems([...items.map((v, idx) => ({ ...v, idx }))]);
  };

  const onChangeMLPSecretName = (idx) => {
    return (e) => {
      items[idx] = { ...items[idx], mlp_secret_name: e };

      setItems((_) =>
        items[items.length - 1].mlp_secret_name &&
        items[items.length - 1].mlp_secret_name.trim()
          ? [...items, { idx: items.length }]
          : [...items],
      );
    };
  };

  const onChangeEnvironmentVariableName = (idx) => {
    return (e) => {
      items[idx] = { ...items[idx], env_var_name: e.target.value };
      setItems((_) => [...items]);
    };
  };

  const { projectId } = useParams();
  const [options, setOptions] = useState([]);

  const [{ data: secrets }] = useMerlinApi(
    `/projects/${projectId}/secrets`,
    {},
    [],
  );

  useEffect(() => {
    if (secrets) {
      const options = [];
      secrets
        .sort((a, b) => (a.name > b.name ? -1 : 1))
        .forEach((secret) => {
          options.push({
            value: secret.name,
            inputDisplay: secret.name,
            textwrap: "truncate",
          });
        });
      setOptions(options);
    }
  }, [secrets]);

  const columns = [
    {
      name: "MLP Secret Name",
      field: "mlp_secret_name",
      width: "45%",
      textOnly: false,
      render: (name, item) => (
        <EuiSuperSelect
          placeholder={"Select MLP secret"}
          compressed={true}
          options={options}
          valueOfSelected={name}
          onChange={onChangeMLPSecretName(item.idx)}
          hasDividers
        />
      ),
    },
    {
      name: "Environment Variable Name",
      field: "env_var_name",
      width: "45%",
      render: (value, item) => (
        <EuiFieldText
          controlOnly
          className="inlineTableInput"
          placeholder="Environment Variable Name"
          value={value || ""}
          onChange={onChangeEnvironmentVariableName(item.idx)}
        />
      ),
    },
    {
      width: "10%",
      actions: [
        {
          render: (item) => {
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
          },
        },
      ],
    },
  ];

  return (
    <Fragment>
      <EuiTitle size="xs">
        <h4>Secrets</h4>
      </EuiTitle>

      <EuiInMemoryTable
        className="Secrets"
        columns={columns}
        items={items}
      />
    </Fragment>
  );
};

SecretsForm.propTypes = {
  variables: PropTypes.array,
  onChange: PropTypes.func,
};
