import { EuiText } from "@elastic/eui";
import { Handle, Position } from "react-flow-renderer";

function PipelineNode({ data }) {
  return (
    <div>
      <Handle type="target" position="left" />
      <div>
        <EuiText title={data.op_type} />
        <hr />
      </div>
      <Handle type="source" position="right" />
    </div>
  );
}

export default PipelineNode;
