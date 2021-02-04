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
import { EuiComboBox } from "@elastic/eui";

export const FeastProjectComboBox = ({
  fullWidth,
  project,
  feastProjects,
  onChange
}) => {
  const [allOptions, setAllOptions] = useState([]);
  useEffect(() => {
    if (allOptions.length === 0 && feastProjects && feastProjects.projects) {
      let options = [];
      feastProjects.projects
        .sort((a, b) => (a > b ? 1 : -1))
        .forEach(project => {
          options.push({ key: project, label: project });
        });
      setAllOptions(options);
    }
  }, [allOptions, feastProjects]);

  const [selectedProjects, setSelectedProjects] = useState([]);
  useEffect(() => {
    setSelectedProjects([{ label: project }]);
  }, [project]);

  const onProjectChange = selectedProjects => {
    setSelectedProjects(selectedProjects);
    selectedProjects &&
      selectedProjects.length === 1 &&
      onChange(selectedProjects[0].label);
  };

  const onProjectCreate = searchValue => {
    const normalizedSearchValue = searchValue.trim();
    setSelectedProjects([{ label: normalizedSearchValue }]);
    onChange(normalizedSearchValue);
  };

  return (
    <EuiComboBox
      fullWidth={fullWidth}
      singleSelection={{ asPlainText: true }}
      isClearable={false}
      placeholder="Project"
      options={allOptions}
      onChange={onProjectChange}
      onCreateOption={onProjectCreate}
      selectedOptions={selectedProjects}
    />
  );
};

FeastProjectComboBox.propTypes = {
  fullWidth: PropTypes.bool,
  project: PropTypes.string,
  feastProjects: PropTypes.object,
  onChange: PropTypes.func
};
