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

import React from "react";
import PropTypes from "prop-types";
import { EuiDescriptionList } from "@elastic/eui";

export const ContainerConfigTable = ({ config: { image, command, args } }) => {
  const items = [
    {
      title: "Image",
      description: image
    }
  ];

  if (command) {
    items.push({
      title: "Command",
      description: command
    });
  }

  if (args) {
    items.push({
      title: "Args",
      description: args
    });
  }

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      columnWidths={[1, 4]}
    />
  );
};

ContainerConfigTable.propTypes = {
  config: PropTypes.object.isRequired
};
