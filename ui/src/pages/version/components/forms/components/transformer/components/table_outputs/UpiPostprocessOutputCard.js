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
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { DraggableHeader } from "../../../DraggableHeader";

export const UpiPostprocessOutputCard = ({ output, onDelete, onChangeHandler, errors = {}, ...props }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  return (
    <EuiPanel>
      <DraggableHeader
        onDelete={onDelete}
        dragHandleProps={props.dragHandleProps}
      />
      <EuiSpacer size="s"/>
      <EuiFlexGroup direction="column" gutterSize="m">
        <EuiFlexItem>
          <EuiToolTip content="UPI Postprocess Output will create output in UPI PredictionValueResponse type">
            <span>
            <EuiText size="s">
              <h4>UPI Postprocess Output</h4>
            </EuiText>
              
            </span>
          </EuiToolTip>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            label={
<<<<<<< HEAD
              <EuiToolTip content="Table name that will be used to populate prediction_result_table in the response to client">
=======
              <EuiToolTip content="Prediction Result Table Name corresponding to the prediction rows provided in the request">
>>>>>>> 46453eb (Change some wording and component position)
                <span>
                Prediction Result Table Name <EuiIcon type="questionInCircle" color="subdued" />
                </span>
              </EuiToolTip>
            }
            isInvalid={!!errors.predictionResultTableName}
            error={errors.predictionResultTableName}
            display="columnCompressed"
            fullWidth>
            <EuiFieldText
              placeholder="Prediction Result Table Name"
              value={output.predictionResultTableName || ""}
              onChange={e =>
                onChange("0.upiPostprocessOutput.predictionResultTableName")(e.target.value)
              }
              name="prediction-result-table-name"
              isInvalid={!!errors.predictionResultTableName}
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
      
    </EuiPanel>
  )
}
