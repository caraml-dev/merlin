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
import { EuiCopy, EuiIcon, EuiLink, EuiText, EuiTextColor } from "@elastic/eui";

export const CopyableUrl = ({ text, iconSize }) => {
  return text ? (
    <EuiCopy textToCopy={text} beforeMessage="Click to copy URL to clipboard">
      {copy => (
        <EuiLink
          onClick={e => {
            e.stopPropagation();
            copy();
          }}
          color="text">
          <EuiText size="s">
            <EuiIcon
              type={"copyClipboard"}
              size={iconSize || "s"}
              style={{ marginRight: "inherit" }}
            />
            &nbsp;{text}
          </EuiText>
        </EuiLink>
      )}
    </EuiCopy>
  ) : (
    <EuiTextColor color="subdued">Not available</EuiTextColor>
  );
};

CopyableUrl.propTypes = {
  text: PropTypes.string.isRequired,
  iconSize: PropTypes.oneOf(["xs", "s", "m", "l", "xl"])
};
