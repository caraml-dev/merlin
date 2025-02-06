import React, { useEffect, useState } from "react";
import { EuiButtonIcon, EuiFieldText, EuiSpacer, EuiSuperSelect } from "@elastic/eui";
import { InMemoryTableForm, useOnChangeHandler } from "@caraml-dev/ui-lib";
import { Panel } from "./Panel";
import { useMerlinApi } from "../../../../../hooks/useMerlinApi";
import { useParams } from "react-router-dom";
import "./SecretsPanel.scss"

export const SecretsPanel = ({
  variables,
  onChangeHandler,
  errors = {}
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
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
            textwrap: "truncate"
          });
        });
      setOptions(options);
    }
  }, [secrets]);

  const items = [
    ...variables.map((v, idx) => ({ idx, ...v })),
    { idx: variables.length }
  ];

  const onDeleteVariable = idx => () => {
    variables.splice(idx, 1);
    onChangeHandler(variables);
  };

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  const columns = [
    {
      name: "MLP Secret Name",
      field: "mlp_secret_name",
      width: "45%",
      render: (name, item) => (
        <EuiSuperSelect
          placeholder={"Select MLP secret"}
          compressed={true}
          options={options}
          valueOfSelected={name}
          onChange={e => onChange(`${item.idx}.mlp_secret_name`)(e)}
          hasDividers
        />
      )
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
          onChange={e => onChange(`${item.idx}.env_var_name`)(e.target.value)}
        />
      )
    },
    {
      width: "10%",
      actions: [
        {
          render: item =>
            item.idx < items.length - 1 ? (
              <EuiButtonIcon
                size="s"
                color="danger"
                iconType="trash"
                onClick={onDeleteVariable(item.idx)}
                aria-label="Remove variable"
              />
            ) : (
              <div />
            )
        }
      ]
    }
  ];

  return (
    <Panel title="Secrets">
      <EuiSpacer size="xs" />
      <InMemoryTableForm
        columns={columns}
        rowProps={getRowProps}
        className={"Secrets"}
        items={items}
        errors={errors}
        renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
      />
    </Panel>
  );
};
