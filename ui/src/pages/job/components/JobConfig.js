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
          <EuiFlexGroup gutterSize="xl">
            <EuiFlexItem>
              <ConfigSectionPanel>
                <SourceConfig job={job} />
              </ConfigSectionPanel>
            </EuiFlexItem>
            <EuiFlexItem>
              <ConfigSectionPanel>
                <SinkConfig job={job} />
              </ConfigSectionPanel>
            </EuiFlexItem>
          </EuiFlexGroup>
        </ConfigSection>
      </EuiFlexItem>

      <EuiFlexItem>
        <ConfigSection title="Runtime Configuration">
          <EuiFlexGroup gutterSize="xl">
            <EuiFlexItem>
              <ConfigSectionPanel>
                <ConfigSectionPanelTitle title="Environment Variables" />
                <EnvVarsConfigTable
                  variables={job.config.env_vars ? job.config.env_vars : []}
                />
              </ConfigSectionPanel>
            </EuiFlexItem>

            <EuiFlexItem>
              <ConfigSectionPanel>
                <ConfigSectionPanelTitle title="Resource Requests" />
                <ResourcesConfigTable
                  resourceRequest={job.config.resource_request}
                />
              </ConfigSectionPanel>
            </EuiFlexItem>
          </EuiFlexGroup>
        </ConfigSection>
      </EuiFlexItem>

      <EuiFlexItem>
        <ConfigSection title="Other Configuration">
          <EuiFlexGroup gutterSize="xl">
            <EuiFlexItem>
              <ConfigSectionPanel>
                <OtherConfig job={job} />
              </ConfigSectionPanel>
            </EuiFlexItem>
          </EuiFlexGroup>
        </ConfigSection>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

export default JobConfig;
