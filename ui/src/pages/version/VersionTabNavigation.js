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

import { EuiIcon } from "@elastic/eui";
import React from "react";
import { useNavigate } from "react-router-dom";
import { TabNavigation } from "../../components/TabNavigation";

export const VersionTabNavigation = ({ endpoint, actions, selectedTab }) => {
  const navigate = useNavigate();
  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
    {
      id: "history",
      name: "History",
    },
    {
      id: "logs",
      name: "Logs",
    },
    {
      id: "monitoring_dashboard_link",
      name: (
        <span>
          Monitoring&nbsp;
          <EuiIcon className="eui-alignBaseline" type="popout" size="s" />
        </span>
      ),
      href: endpoint.monitoring_url,
      target: "_blank",
    },
  ];

  return (
    <TabNavigation
      tabs={tabs}
      actions={actions}
      selectedTab={selectedTab}
      navigate={navigate}
    />
  );
};
