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

import React, { useEffect, useState } from "react";
import PropTypes from "prop-types";
import {
  EuiButtonIcon,
  EuiComboBox,
  EuiFieldText,
  EuiIcon,
  EuiInMemoryTable,
  EuiToolTip
} from "@elastic/eui";

const getSelectedOption = value => (value ? [{ label: value }] : []);

export const FeastEntities = ({ entities, feastEntities, onChange }) => {
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
          jsonPath: item.jsonPath
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
      items[items.length - 1].name &&
      items[items.length - 1].valueType &&
      items[items.length - 1].jsonPath &&
      items[items.length - 1].fieldType
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
        />
      )
    },
    {
      name: "Value Type",
      field: "valueType",
      width: "25%",
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
        />
      )
    },
    {
      name: "Field Type",
      field: "fieldType",
      width: "30%",
      render: (value, item) => (
        <EuiComboBox
          fullWidth
          singleSelection={{ asPlainText: true }}
          isClearable={false}
          placeholder="Field Type"
          options={[{ label: "JSONPath" }, { label: "UDF" }]}
          onChange={onEntityChange(item.idx, "fieldType")}
          onCreateOption={searchValue =>
            onEntityCreate(searchValue, item.idx, "fieldType")
          }
          selectedOptions={getSelectedOption(value)}
        />
      )
    },
    {
      name: (
        <EuiToolTip content="Specify the JSONPath/UDF syntax to extract entity value from the request payload.">
          <span>
            Field <EuiIcon type="questionInCircle" color="subdued" />
          </span>
        </EuiToolTip>
      ),
      field: "jsonPath",
      width: "20%",
      render: (value, item) => (
        <EuiFieldText
          controlOnly
          fullWidth
          placeholder="Field"
          value={value || ""}
          onChange={onChangeRow(item.idx, "jsonPath")}
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

  return <EuiInMemoryTable columns={columns} items={items} hasActions={true} />;
};

FeastEntities.propTypes = {
  entities: PropTypes.array,
  feastEntities: PropTypes.object,
  onChange: PropTypes.func
};
