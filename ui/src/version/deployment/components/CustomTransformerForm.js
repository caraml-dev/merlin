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
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiSuperSelect,
  EuiToolTip
} from "@elastic/eui";
import { appConfig } from "../../../config";

const extractRegistry = (image, registries) => {
  if (image) {
    const registry = registries.find(o => image.startsWith(o.value));
    if (registry && registry.value) {
      image = image.substr(registry.value.length);
      image && image.startsWith("/") && (image = image.substr(1));
      return [registry.value, image];
    }
  }
  return ["docker-hub", image];
};

const dockerRegistryDisplay = registry => (
  <EuiFlexGroup alignItems="center" gutterSize="s">
    <EuiFlexItem grow={1}>
      <EuiIcon type="logoDocker" size="m" />
    </EuiFlexItem>
    <EuiFlexItem grow={9}>{registry}</EuiFlexItem>
  </EuiFlexGroup>
);

export const CustomTransformerForm = ({ transformer, onTransformerChange }) => {
  const [dockerRegistries, setDockerRegistries] = useState([
    {
      value: "docker-hub",
      inputDisplay: dockerRegistryDisplay("Docker Hub"),
      dropdownDisplay: dockerRegistryDisplay("Docker Hub")
    }
  ]);

  useEffect(
    () => {
      if (appConfig.dockerRegistries) {
        setDockerRegistries([
          ...dockerRegistries,
          ...appConfig.dockerRegistries.map(registry => ({
            value: registry,
            inputDisplay: dockerRegistryDisplay(registry),
            dropdownDisplay: dockerRegistryDisplay(registry)
          }))
        ]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [appConfig.dockerRegistries]
  );

  const [registry, image] = extractRegistry(
    transformer.image,
    dockerRegistries
  );

  const setValue = (field, value) =>
    onTransformerChange({
      ...transformer,
      [field]: value
    });

  const onRegistryChange = value => {
    setValue("image", value !== "docker-hub" ? `${value}/${image}` : image);
  };

  const onImageChange = value => {
    setValue(
      "image",
      registry !== "docker-hub" ? `${registry}/${value}` : value
    );
  };

  return (
    <>
      <EuiFormRow fullWidth label="Docker Image*">
        <EuiFlexGroup gutterSize="s">
          <EuiFlexItem grow={4}>
            <EuiSuperSelect
              fullWidth
              options={dockerRegistries}
              valueOfSelected={registry}
              onChange={value => onRegistryChange(value)}
              itemLayoutAlign="top"
              hasDividers
            />
          </EuiFlexItem>
          <EuiFlexItem grow={6}>
            <EuiFieldText
              fullWidth
              placeholder="Docker image name and tag"
              value={image}
              onChange={e => onImageChange(e.target.value)}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFormRow>

      <EuiFormRow
        fullWidth
        label={
          <EuiToolTip content="Specify the command to be executed. The Docker image's ENTRYPOINT is used if this is not provided.">
            <span>
              Command <EuiIcon type="questionInCircle" color="subdued" />
            </span>
          </EuiToolTip>
        }
        display="columnCompressed">
        <EuiFieldText
          fullWidth
          value={transformer.command || ""}
          onChange={e => setValue("command", e.target.value)}
          name="command"
        />
      </EuiFormRow>

      <EuiFormRow
        fullWidth
        label={
          <EuiToolTip content="Specify the arguments to the command above. The Docker image's CMD is used if this is not provided.">
            <span>
              Args <EuiIcon type="questionInCircle" color="subdued" />
            </span>
          </EuiToolTip>
        }
        display="columnCompressed">
        <EuiFieldText
          fullWidth
          value={transformer.args || ""}
          onChange={e => setValue("args", e.target.value)}
          name="args"
        />
      </EuiFormRow>
    </>
  );
};

CustomTransformerForm.propTypes = {
  transformer: PropTypes.object,
  onTransformerChange: PropTypes.func
};
