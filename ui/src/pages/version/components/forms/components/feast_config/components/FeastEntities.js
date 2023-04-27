import React, { useEffect, useState } from "react";
import {
  EuiButtonIcon,
  EuiComboBox,
  EuiFieldText,
  EuiIcon,
  EuiToolTip
} from "@elastic/eui";
import { get, InMemoryTableForm } from "@caraml-dev/ui-lib";
import { JsonPathConfigInput } from "../../transformer/JsonPathConfigInput";
import "../../transformer/RowCell.scss";

const getSelectedOption = value => (value ? [{ label: value }] : []);

export const FeastEntities = ({
  entities,
  feastEntities,
  onChange,
  errors = {}
}) => {
  const [allEntities, setAllEntities] = useState([]);
  const [allValueTypes, setAllValueTypes] = useState([]);
  useEffect(() => {
    if (feastEntities && feastEntities.entities) {
      const allEntities = [];
      const allValueTypes = [];
      let entitiesMap = new Map();
      let valueTypesMap = new Map();
      feastEntities.entities.forEach(entity => {
        if (entity.spec.name) {
          if (!entitiesMap.has(entity.spec.name)) {
            allEntities.push({
              key: entity.spec.name,
              label: entity.spec.name,
              spec: entity.spec
            });
            entitiesMap.set(entity.spec.name, 1);
          }

          if (!valueTypesMap.has(entity.spec.valueType)) {
            allValueTypes.push({
              key: entity.spec.valueType,
              label: entity.spec.valueType
            });
            valueTypesMap.set(entity.spec.valueType, 1);
          }
        }
      });

      allEntities.sort((a, b) => (a.label > b.label ? 1 : -1));
      setAllEntities(allEntities);

      allValueTypes.sort((a, b) => (a.label > b.label ? 1 : -1));
      setAllValueTypes(allValueTypes);
    }
  }, [feastEntities]);

  const [items, setItems] = useState([
    ...entities.map((v, idx) => ({ idx, ...v })),
    { idx: entities.length }
  ]);

  useEffect(
    () => {
      const updatedItems = [
        ...entities.map((v, idx) => ({ idx, ...v })),
        { idx: entities.length }
      ];

      if (JSON.stringify(items) !== JSON.stringify(updatedItems)) {
        setItems(updatedItems);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [entities]
  );

  useEffect(
    () => {
      if (items.length > 1) {
        const updatedItems = items.slice(0, items.length - 1).map(item => ({
          name: item.name,
          valueType: item.valueType,
          fieldType: item.fieldType,
          field: item.field
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
      items[items.length - 1].name ||
      items[items.length - 1].valueType ||
      items[items.length - 1].field ||
      items[items.length - 1].fieldType
        ? [...items, { idx: items.length }]
        : [...items]
    );
  };

  const onChangeRow = (idx, field) => {
    return e => onVariableChange(idx, field, e.target.value);
  };

  const onVariableChange = (idx, field, value) => {
    items[idx] = {
      ...items[idx],
      [field]: value
    };
    updateItems(items);
  };

  const onEntityChange = (idx, field) => {
    return e => {
      items[idx] = {
        ...items[idx],
        [field]: e[0] ? e[0].label : ""
      };
      if (field === "name") {
        items[idx] = {
          ...items[idx],
          valueType: e[0] && e[0].spec ? e[0].spec.valueType : ""
        };
      }
      if (field === "fieldType") {
        //flatten value type for non-jsonpath type
        if (
          items[idx].fieldType !== "JSONPath" &&
          typeof items[idx].field === "object"
        ) {
          items[idx].field = items[idx].field.jsonPath;
        }

        switch (items[idx].fieldType) {
          case "JSONPath":
            if (items[idx].field && items[idx].field.jsonPath === undefined) {
              items[idx].field = { jsonPath: items[idx].field };
            }
            break;
          default:
            break;
        }
      }
      updateItems(items);
    };
  };

  const onEntityCreate = (searchValue, idx, field) => {
    const normalizedSearchValue = searchValue.trim();
    items[idx] = {
      ...items[idx],
      [field]: normalizedSearchValue
    };
    updateItems(items);
  };

  const columns = [
    {
      name: "Entity",
      field: "name",
      width: "30%",
      render: (value, item) => (
        <EuiComboBox
          fullWidth
          singleSelection={{ asPlainText: true }}
          isClearable={false}
          placeholder="Entity"
          options={allEntities}
          onChange={onEntityChange(item.idx, "name")}
          onCreateOption={searchValue =>
            onEntityCreate(searchValue, item.idx, "name")
          }
          selectedOptions={getSelectedOption(value)}
          isInvalid={!!get(errors, `${item.idx}.name`)}
        />
      )
    },
    {
      name: "Entity Value Type",
      field: "valueType",
      width: "20%",
      render: (value, item) => (
        <EuiComboBox
          fullWidth
          singleSelection={{ asPlainText: true }}
          isClearable={false}
          placeholder="Value Type"
          options={allValueTypes}
          onChange={onEntityChange(item.idx, "valueType")}
          onCreateOption={searchValue =>
            onEntityCreate(searchValue, item.idx, "valueType")
          }
          selectedOptions={getSelectedOption(value)}
          isInvalid={!!get(errors, `${item.idx}.valueType`)}
        />
      )
    },
    {
      name: "Input Type",
      field: "fieldType",
      width: "20%",
      render: (value, item) => (
        <EuiComboBox
          fullWidth
          singleSelection={{ asPlainText: true }}
          isClearable={false}
          placeholder="Field Type"
          options={[{ label: "JSONPath" }, { label: "Expression" }]}
          onChange={onEntityChange(item.idx, "fieldType")}
          onCreateOption={searchValue =>
            onEntityCreate(searchValue, item.idx, "fieldType")
          }
          selectedOptions={getSelectedOption(value)}
          isInvalid={!!get(errors, `${item.idx}.fieldType`)}
        />
      )
    },
    {
      name: (
        <EuiToolTip content="Specify the JSONPath/Expression syntax to extract entity value from the request payload">
          <span>
            Input Value <EuiIcon type="questionInCircle" color="subdued" />
          </span>
        </EuiToolTip>
      ),
      field: "field",
      width: "30%",
      render: (value, item) => {
        if (item.fieldType === "JSONPath") {
          return (
            <JsonPathConfigInput
              jsonPathConfig={value}
              identifier={`entities-${item.idx}`}
              onChangeHandler={val => onVariableChange(item.idx, "field", val)}
            />
          );
        }
        return (
          <EuiFieldText
            controlOnly
            fullWidth
            placeholder="Value"
            value={value || ""}
            onChange={onChangeRow(item.idx, "field")}
            isInvalid={!!get(errors, `${item.idx}.field`)}
          />
        );
      }
    },
    {
      width: "5%",
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
