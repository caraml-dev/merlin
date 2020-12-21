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
import { EuiCode, EuiComboBox, EuiFormRow } from "@elastic/eui";
import PropTypes from "prop-types";

export const FeatureComboBox = ({ options, onChange }) => {
  const [features, setFeatures] = useState([]);
  const [selectedFeatures, setSelectedFeatures] = useState(options);

  useEffect(
    () => {
      if (selectedFeatures.length > 0) {
        const inputFeatures = [];
        selectedFeatures.forEach(feature => {
          inputFeatures.push(feature.label);
        });
        onChange(inputFeatures);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [selectedFeatures]
  );

  const onFeaturesChange = features => {
    setSelectedFeatures(features);
  };

  const onFeatureCreate = (searchValue, flattenedOptions = []) => {
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
      setFeatures([...features, newOption]);
    }

    // Select the option.
    setSelectedFeatures([...selectedFeatures, newOption]);
  };

  return (
    <EuiFormRow
      fullWidth
      label="Data Source Features *"
      helpText={
        <p>
          Specify the table columns that will be used as input features. Use{" "}
          <EuiCode>,</EuiCode> as delimiter.
        </p>
      }>
      <EuiComboBox
        fullWidth
        noSuggestions
        delimiter=","
        isClearable={false}
        options={features}
        onChange={onFeaturesChange}
        onCreateOption={onFeatureCreate}
        selectedOptions={selectedFeatures}
      />
    </EuiFormRow>
  );
};

FeatureComboBox.propTypes = {
  options: PropTypes.array,
  onChange: PropTypes.func
};
