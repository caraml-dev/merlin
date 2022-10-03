import {
  EuiDragDropContext,
  EuiDraggable,
  EuiDroppable,
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer
} from "@elastic/eui";
import { Panel } from "../Panel";
import {
  JsonOutputCard
} from "./components/table_outputs/JsonOutputCard";
import {
  UpiPreprocessOutputCard
} from "./components/table_outputs/UpiPreprocessOutputCard";
import {
  UpiPostprocessOutputCard
} from "./components/table_outputs/UpiPostprocessOutputCard";
import { Output } from "../../../../../../services/transformer/TransformerConfig";
import { AddButton } from "./components/AddButton";


export const OutputPanel = ({ outputs = [], protocol, pipelineStage, onChangeHandler, errors = {} }) => {
  const onAddOutput = () => {
    let output = new Output(protocol, pipelineStage)
    onChangeHandler([...outputs,  output]);
  };
  const onDeleteOutputIdx = (idx) => {
    outputs.splice(idx, 1);
    onChangeHandler([...outputs])
  }
  return (
    <Panel title="Output" contentWidth="92%">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="s">
        <EuiDragDropContext>
          <EuiDroppable droppableId="OUTPUTS_DROPPABLE_AREA" spacing="m">
          { outputs.map((output, idx) => (
            <EuiDraggable
                key={`output-${idx}`}
                index={idx}
<<<<<<< HEAD
                draggableId={`output-${idx}`}
=======
                draggableId={`${idx}`}
>>>>>>> cf45b71 (Add output and autoload standard transformer UI components)
                customDragHandle={true}
                disableInteractiveElementBlocking>
              {
                provided => (
                  <EuiFlexItem>
                    {output.jsonOutput && (
                        <JsonOutputCard 
                          jsonOutput={output.jsonOutput}
                          onChangeHandler={onChangeHandler}
                          errors={errors}
                          onDelete={() => onDeleteOutputIdx(idx)}
                          dragHandleProps={provided.dragHandleProps}
                        />
                      )
                    }
                    {
                      output.upiPreprocessOutput && (
                        <UpiPreprocessOutputCard
                          output={output.upiPreprocessOutput}
                          onDelete={() => onDeleteOutputIdx(idx)}
                          onChangeHandler={onChangeHandler}
                          errors={errors}
                          dragHandleProps={provided.dragHandleProps}
                        />
                      )
                    }
                    {
                      output.upiPostprocessOutput && (
                        <UpiPostprocessOutputCard
                          output={output.upiPostprocessOutput}
                          onDelete={() => onDeleteOutputIdx(idx)}
                          onChangeHandler={onChangeHandler}
                          errors={errors}
                          dragHandleProps={provided.dragHandleProps}
                        />
                      )
                    }
                  </EuiFlexItem>
                )
              }
            </EuiDraggable>
          ))
          }  
          </EuiDroppable>
        </EuiDragDropContext>
        {outputs.length === 0 && (
          <EuiFlexItem>
            <AddButton
              title="+ Add Output"
              description="Add Output operation"
              titleSize="xs"
              onClick={() => onAddOutput()} />
          </EuiFlexItem>)
        }
      </EuiFlexGroup>
    </Panel>
  );
};
