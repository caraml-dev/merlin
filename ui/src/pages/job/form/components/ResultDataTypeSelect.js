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

import React, { useState } from "react";
import { EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import PropTypes from "prop-types";

export const ResultDataTypeSelect = ({
  resultType,
  resultItemType,
  setResultType,
  setResultItemType,
}) => {
  const options = [
    {
      value: "DOUBLE",
      inputDisplay: "Double",
      dropdownDisplay: "Double",
    },
    {
      value: "FLOAT",
      inputDisplay: "Float",
      dropdownDisplay: "Float",
    },
    {
      value: "INTEGER",
      inputDisplay: "Integer",
      dropdownDisplay: "Integer",
    },
    {
      value: "LONG",
      inputDisplay: "Long",
      dropdownDisplay: "Long",
    },
    {
      value: "STRING",
      inputDisplay: "String",
      dropdownDisplay: "String",
    },
    {
      value: "ARRAY OF DOUBLE",
      inputDisplay: "Array of double",
      dropdownDisplay: "Array of double",
    },
    {
      value: "ARRAY OF FLOAT",
      inputDisplay: "Array of float",
      dropdownDisplay: "Array of float",
    },
    {
      value: "ARRAY OF INTEGER",
      inputDisplay: "Array of integer",
      dropdownDisplay: "Array of integer",
    },
    {
      value: "ARRAY OF LONG",
      inputDisplay: "Array of long",
      dropdownDisplay: "Array of long",
    },
    {
      value: "ARRAY OF STRING",
      inputDisplay: "Array of string",
      dropdownDisplay: "Array of string",
    },
  ];

  const [selected, setSelected] = useState(
    resultType && resultItemType && resultType.includes("ARRAY")
      ? `ARRAY OF ${resultItemType}`
      : resultType,
  );

  const onChange = (value) => {
    setSelected(value);

    if (value.includes("ARRAY")) {
      const parts = value.split(" ");
      setResultType("ARRAY");
      setResultItemType(parts[2]);
    } else {
      setResultType(value);
    }
  };

  return (
    <EuiFormRow
      fullWidth
      label="Result Data Type *"
      helpText="Choose the data type of your batch job result"
    >
      <EuiSuperSelect
        fullWidth
        options={options}
        valueOfSelected={selected}
        onChange={onChange}
        hasDividers
      />
    </EuiFormRow>
  );
};

ResultDataTypeSelect.propTypes = {
  resultType: PropTypes.string,
  resultItemType: PropTypes.string,
  setResultType: PropTypes.func,
  setResultItemType: PropTypes.func,
};
