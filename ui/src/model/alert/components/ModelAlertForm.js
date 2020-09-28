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

import React, { useState, useEffect } from "react";
import {
  EuiComboBox,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiTitle,
  EuiToolTip
} from "@elastic/eui";
import { useMerlinApi } from "../../../hooks/useMerlinApi";
import { ModelAlertRule } from "./ModelAlertRule";
import { ModelAlertRulePercentile } from "./ModelAlertRulePercentile";

export const ModelAlertForm = ({ request, setRequest }) => {
  const [{ data: teams }] = useMerlinApi(
    `/alerts/teams`,
    { muteError: true },
    [],
    true
  );

  const [teamOptions, setTeamOptions] = useState([]);
  useEffect(() => {
    teams &&
      setTeamOptions(
        teams
          .sort((a, b) => (a > b ? 1 : -1))
          .map((team, idx) => ({ key: idx, label: team }))
      );
  }, [teams, setTeamOptions]);

  const [selectedTeams, setSelectedTeams] = useState([]);

  const onTeamChange = selectedTeams => {
    setSelectedTeams(selectedTeams);
    selectedTeams &&
      selectedTeams.length === 1 &&
      setRequest({ ...request, team_name: selectedTeams[0].label });
  };

  const onTeamCreate = searchValue => {
    const normalizedSearchValue = searchValue.trim().toLowerCase();
    setSelectedTeams([{ label: normalizedSearchValue }]);
    setRequest({ ...request, team_name: normalizedSearchValue });
  };

  useEffect(() => {
    request.team_name && setSelectedTeams([{ label: request.team_name }]);
  }, [request.team_name, setSelectedTeams]);

  return (
    <EuiPanel grow={false}>
      <EuiTitle size="xs">
        <h4>Alerts</h4>
      </EuiTitle>

      <EuiSpacer size="l" />
      <EuiFlexGroup>
        <EuiFlexItem grow={1}>
          <EuiFormRow
            fullWidth
            display="columnCompressed"
            label={
              <EuiToolTip content="Specify the team responsible for the alert. The notification will be sent to the team's Slack alerting channel.">
                <span>
                  Team Name <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            // TODO: Provide correct link to team naming guide.
            helpText={
              <EuiText size="xs">
                Not sure about your team name? Please check{" "}
                <a href="/" target="_blank" rel="noopener noreferrer">
                  this guide
                </a>
                .
              </EuiText>
            }>
            <EuiComboBox
              singleSelection={{ asPlainText: true }}
              isClearable={false}
              options={teamOptions}
              onChange={onTeamChange}
              onCreateOption={onTeamCreate}
              selectedOptions={selectedTeams}
              placeholder="e.g. fraud, gofood, gopay"
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>

      <ModelAlertRule
        title="Throughput"
        name="throughput"
        description="Send alert if throughput is lower than threshold."
        request={request}
        setRequest={setRequest}
      />

      <ModelAlertRulePercentile
        title="Latency"
        name="latency"
        description="Send alert if request duration is higher than threshold."
        unit="ms"
        request={request}
        setRequest={setRequest}
      />

      <ModelAlertRule
        title="Error Rate"
        name="error_rate"
        description="Send alert if error rate is higher than threshold."
        unit="%"
        max={100}
        request={request}
        setRequest={setRequest}
      />

      <ModelAlertRule
        title="CPU"
        name="cpu"
        description="Send alert if CPU usage is higher than threshold."
        unit="%"
        max={100}
        request={request}
        setRequest={setRequest}
      />

      <ModelAlertRule
        title="Memory"
        name="memory"
        description="Send alert if memory usage is higher than threshold."
        unit="%"
        max={100}
        request={request}
        setRequest={setRequest}
      />
    </EuiPanel>
  );
};
