import {
  EuiFlexGroup,
  EuiPanel,
  EuiFlexItem,
  EuiFormRow,
  EuiFieldText,
  EuiToolTip,
  EuiText,
  EuiSpacer,
  EuiIcon

} from "@elastic/eui";
import {
  get, 
  useOnChangeHandler
} from "@gojek/mlp-ui";
import {ColumnsComboBox} from "../table_operations/ColumnsComboBox"
import { DraggableHeader } from "../../../DraggableHeader";


export const UpiPreprocessOutputCard = ({ output, onDelete, onChangeHandler, errors = {}, ...props }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="m">
        <EuiFlexItem>
          <EuiToolTip content="UPI Preprocess Output will create output in UPI PredictionValueRequest type">
            <span>
            <EuiText size="s">
              <h4>UPI Preprocess Output</h4>
            </EuiText>
              
            </span>
          </EuiToolTip>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            label={
              <EuiToolTip content="Prediction Table Name contains instances to be predicted. This should contain all preprocessed feature that model use to perform prediction">
                <span>
                Prediction Table Name <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            isInvalid={!!errors.predictionTableName}
            error={errors.predictionTableName}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Prediction Table Name"
              value={output.predictionTableName || ""}
              onChange={e =>
                onChange("0.upiPreprocessOutput.predictionTableName")(e.target.value)
              }
              name="prediction-table-name"
              isInvalid={!!errors.predictionTableName}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <ColumnsComboBox
            columns={output.transformerInputTableNames || []}
            onChange={onChange("0.upiPreprocessOutput.transformerInputTableNames")}
            title="Transformer input table names"
            description={
              <p>
                List of transformer input table name
              </p>
            }
            errors={get(errors, "transformerInputTableNames")}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
      
    </EuiPanel>
  )
}
