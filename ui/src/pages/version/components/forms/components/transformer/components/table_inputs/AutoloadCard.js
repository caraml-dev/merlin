import {
  EuiFlexGroup,
  EuiPanel,
  EuiFlexItem,
  EuiSpacer,
  EuiToolTip,
  EuiText
  } from "@elastic/eui";
import { get } from "@gojek/mlp-ui";
import { useOnChangeHandler } from "@gojek/mlp-ui";
import { ColumnsComboBox} from "../table_operations/ColumnsComboBox"
import { DraggableHeader } from "../../../DraggableHeader";
  
export const AutoloadCard = ({ 
    autoload={}, 
    onDelete, 
    onChangeHandler, 
    errors = {}, 
    ...props }) => {
    
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
            <EuiToolTip content="Autoload will load tables or variables from request/model response">
              <span>
              <EuiText size="s">
                <h4>Autoload</h4>
              </EuiText>
                
              </span>
            </EuiToolTip>
          </EuiFlexItem>
          <EuiFlexItem>
            <ColumnsComboBox
              columns={autoload.tableNames || []}
              onChange={onChange("tableNames")}
              title="Table names"
              description={
                <p>
                  List of table name that must be loaded from request/response and can be used for other operations
                </p>
              }
              errors={get(errors, "tableNames")}
            />
          </EuiFlexItem>
          <EuiFlexItem>
            <ColumnsComboBox
              columns={autoload.variableNames || []}
              onChange={onChange("variableNames")}
              title="Variable names"
              description={
                <p>
                  List of variable name that must be loaded from request/response and can be used for other operations
                </p>
              }
              errors={get(errors, "variableNames")}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
        
      </EuiPanel>
    )
  }
  