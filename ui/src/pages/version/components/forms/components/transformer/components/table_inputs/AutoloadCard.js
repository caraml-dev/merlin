import {
  EuiFlexGroup,
  EuiPanel,
  EuiFlexItem,
  EuiSpacer,
  EuiToolTip,
  EuiText
  } from "@elastic/eui";
import { get } from "@caraml-dev/ui-lib";
import { useOnChangeHandler } from "@caraml-dev/ui-lib";
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
            <EuiToolTip content="UPI Autoload loads tables or variables from UPI payload into standard transformer registry">
              <span>
              <EuiText size="s">
                <h4>UPI Autoload</h4>
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
                  List of table names that will be loaded from request/response and can be used for other operations in standard transformer
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
                  List of variable names that will be loaded from request/response and can be used for other operations in standard transformer
                </p>
              }
              errors={get(errors, "variableNames")}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
        
      </EuiPanel>
    )
  }
  