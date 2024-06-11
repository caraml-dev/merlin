import { EuiCodeBlock } from "@elastic/eui";
import { ConfigSection, ConfigSectionPanel } from "../../../components/section";

const JobErrorMessage = ({ error }) => {
  return (
    <ConfigSection title="Job Error Message">
      <ConfigSectionPanel>
        <EuiCodeBlock isCopyable overflowHeight={900}>
          {error}
        </EuiCodeBlock>
      </ConfigSectionPanel>
    </ConfigSection>
  );
};

export default JobErrorMessage;
