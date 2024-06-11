import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { EnvVarsConfigTable } from "../../../components/EnvVarsConfigTable";
import {
  ConfigSection,
  ConfigSectionPanel,
  ConfigSectionPanelTitle,
} from "../../../components/section";
import OtherConfig from "./OtherConfig";
import { ResourcesConfigTable } from "./ResourcesConfigTable";
import SinkConfig from "./SinkConfig";
import SourceConfig from "./SourceConfig";

const JobConfig = ({ job }) => {
  return (
    <EuiFlexGroup direction="column" gutterSize="l">
      <EuiFlexItem>
        <ConfigSection title="Data Configuration">
          <ConfigSectionPanel>
            <EuiFlexGroup gutterSize="xl">
              <EuiFlexItem>
                <SourceConfig job={job} />
              </EuiFlexItem>

              <EuiFlexItem>
                <SinkConfig job={job} />
              </EuiFlexItem>
            </EuiFlexGroup>
          </ConfigSectionPanel>
        </ConfigSection>
      </EuiFlexItem>

      <EuiFlexItem>
        <ConfigSection title="Runtime Configuration">
          <ConfigSectionPanel>
            <EuiFlexGroup gutterSize="xl">
              <EuiFlexItem>
                <ConfigSectionPanelTitle title="Environment Variables" />
                <EnvVarsConfigTable
                  variables={job.config.env_vars ? job.config.env_vars : []}
                />
              </EuiFlexItem>

              <EuiFlexItem>
                <ConfigSectionPanelTitle title="Resource Requests" />
                <ResourcesConfigTable
                  resourceRequest={job.config.resource_request}
                />
              </EuiFlexItem>
            </EuiFlexGroup>
          </ConfigSectionPanel>
        </ConfigSection>
      </EuiFlexItem>

      <EuiFlexItem>
        <ConfigSection title="Other Configuration">
          <ConfigSectionPanel>
            <EuiFlexGroup gutterSize="xl">
              <EuiFlexItem>
                <OtherConfig job={job} />
              </EuiFlexItem>
            </EuiFlexGroup>
          </ConfigSectionPanel>
        </ConfigSection>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

export default JobConfig;
