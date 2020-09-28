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
import { EuiComboBox, EuiFormRow } from "@elastic/eui";

export const ClusteredFieldsComboBox = ({ options, onChange }) => {
  const [fields, setFields] = useState([]);
  const [selectedFields, setSelectedFields] = useState(options);

  useEffect(
    () => {
      if (selectedFields.length > 0) {
        const fields = [];
        selectedFields.forEach(feature => {
          fields.push(feature.label);
        });
        onChange(fields.join());
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [selectedFields]
  );

  const onFieldsChange = fields => {
    setSelectedFields(fields);
  };

  const onFieldCreate = (searchValue, flattenedOptions = []) => {
    const normalizedSearchValue = searchValue.trim().toLowerCase();
    if (!normalizedSearchValue) {
      return;
    }

    const newOption = {
      label: searchValue.trim()
    };

    // Create the option if it doesn't exist.
    if (
      flattenedOptions.findIndex(
        option => option.label.trim().toLowerCase() === normalizedSearchValue
      ) === -1
    ) {
      setFields([...fields, newOption]);
    }

    // Select the option.
    setSelectedFields([...selectedFields, newOption]);
  };

  return (
    <EuiFormRow
      fullWidth
      label="Clustered Fields"
      helpText={
        <p>
          List of non-repeated, top level columns to cluster partitioned tables.
        </p>
      }>
      <EuiComboBox
        fullWidth
        noSuggestions
        delimiter=","
        isClearable={false}
        options={fields}
        onChange={onFieldsChange}
        onCreateOption={onFieldCreate}
        selectedOptions={selectedFields}
      />
    </EuiFormRow>
  );
};
