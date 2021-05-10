import React, { useEffect, useState } from "react";
import { EuiButtonIcon, EuiComboBox, EuiFieldText } from "@elastic/eui";
import { get, InMemoryTableForm } from "@gojek/mlp-ui";

const getSelectedOption = value => (value ? [{ label: value }] : []);

export const FeastFeatures = ({
  features,
  feastFeatureTables,
  onChange,
  errors = {}
}) => {
  const [allOptions, setAllOptions] = useState([]);
  useEffect(
    () => {
      if (feastFeatureTables && feastFeatureTables.tables) {
        let groups = [];
        feastFeatureTables.tables
          .sort((a, b) => (a.spec.name > b.spec.name ? 1 : -1))
          .forEach(table => {
            let options = [];
            table.spec.features
              .sort((a, b) => (a.name > b.name ? 1 : -1))
              .forEach(feature => {
                const featureFullname = table.spec.name + ":" + feature.name;
                options.push({
                  key: featureFullname,
                  label: feature.name,
                  value: featureFullname,
                  feature: feature
                });
              });
            let group = {
              label: table.spec.name,
              options: options
            };
            groups.push(group);
          });
        setAllOptions(groups);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [feastFeatureTables]
  );

  const [items, setItems] = useState([
    ...features.map((v, idx) => ({ idx, ...v })),
    { idx: features.length }
  ]);

  useEffect(
    () => {
      const updatedItems = [
        ...features.map((v, idx) => ({ idx, ...v })),
        { idx: features.length }
      ];

      if (JSON.stringify(items) !== JSON.stringify(updatedItems)) {
        setItems(updatedItems);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [features]
  );

  useEffect(
    () => {
      if (items.length > 1) {
        const updatedItems = items.slice(0, items.length - 1).map(item => ({
          name: item.name,
          valueType: item.valueType,
          defaultValue: item.defaultValue
        }));
        onChange(updatedItems);
      } else {
        onChange([]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [items]
  );

  const removeRow = idx => {
    items.splice(idx, 1);
    setItems([...items.map((v, idx) => ({ ...v, idx }))]);
  };

  const updateItems = items => {
    setItems(_ =>
      items[items.length - 1].name || items[items.length - 1].defaultValue
        ? [...items, { idx: items.length }]
        : [...items]
    );
  };

  const onChangeRow = (idx, field) => {
    return e => {
      items[idx] = { ...items[idx], [field]: e.target.value };
      updateItems(items);
    };
  };

  const onFeatureChange = idx => {
    return e => {
      items[idx] = {
        ...items[idx],
        name: e[0] ? e[0].value : "",
        valueType: e[0] ? e[0].feature.valueType : ""
      };
      updateItems(items);
    };
  };

  const onFeatureCreate = (searchValue, idx, field) => {
    const normalizedSearchValue = searchValue.trim();
    items[idx] = {
      ...items[idx],
      [field]: normalizedSearchValue
    };
    updateItems(items);
  };

  const columns = [
    {
      name: "Feature Name",
      field: "name",
      width: "60%",
      render: (value, item) => (
        <EuiComboBox
          fullWidth
          singleSelection={{ asPlainText: true }}
          isClearable={false}
          placeholder="Feature Name"
          options={allOptions}
          onChange={onFeatureChange(item.idx, "name")}
          onCreateOption={searchValue =>
            onFeatureCreate(searchValue, item.idx, "name")
          }
          selectedOptions={getSelectedOption(value)}
          isInvalid={!!get(errors, `${item.idx}.name`)}
        />
      )
    },
    {
      name: "Default Value",
      field: "defaultValue",
      width: "30%",
      render: (value, item) => (
        <EuiFieldText
          fullWidth
          controlOnly
          placeholder="Default Value"
          value={value || ""}
          onChange={onChangeRow(item.idx, "defaultValue")}
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

  const getRowProps = item => {
    const { idx } = item;
    const isInvalid = !!errors[idx];
    return {
      className: isInvalid ? "euiTableRow--isInvalid" : "",
      "data-test-subj": `row-${idx}`
    };
  };

  return (
    <InMemoryTableForm
      columns={columns}
      rowProps={getRowProps}
      items={items}
      hasActions={true}
      errors={errors}
      renderErrorHeader={key => `Row ${parseInt(key) + 1}`}
      className={""}
    />
  );
};
