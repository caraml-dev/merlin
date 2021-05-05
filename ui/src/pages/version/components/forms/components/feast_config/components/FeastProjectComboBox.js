import React, { useEffect, useState } from "react";
import { EuiComboBox } from "@elastic/eui";

export const FeastProjectComboBox = ({ project, feastProjects, onChange }) => {
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
    if (project !== "") {
      setSelectedProjects([{ label: project }]);
    }
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
      fullWidth
      singleSelection={{ asPlainText: true }}
      isClearable={false}
      placeholder="Select Feast project"
      options={allOptions}
      onChange={onProjectChange}
      onCreateOption={onProjectCreate}
      selectedOptions={selectedProjects}
    />
  );
};
