import {
  EuiAccordion,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiSideNav,
  EuiText,
  EuiTextColor,
  EuiTitle
} from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import React, { useContext, useState } from "react";
import { PipelineSidebarPanel } from "../components/transformer/PipelineSidebarPanel";
import { PipelineStage } from "../components/transformer/PipelineStage";
import { TransformerSimulation } from "../components/transformer/TransformerSimulation";
import "./Pipeline.scss";

export const StandardTransformerStep = () => {
  const {
    data: {
      transformer: {
        config: {
          transformerConfig: {
            preprocess: { preValues },
            postprocess: { postValues }
          }
        }
      }
    },
    onChangeHandler
  } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);
  const [isSideNavOpenOnMobile, setisSideNavOpenOnMobile] = useState(false);
  const [selectedItemName, setSelectedItem] = useState("Pre-Processing");

  const toggleOpenOnMobile = () => {
    setisSideNavOpenOnMobile(!isSideNavOpenOnMobile);
  };

  const selectItem = name => {
    setSelectedItem(name);
  };

  const createItem = (name, id, data = {}) => {
    return {
      id: id,
      name,
      isSelected: selectedItemName === id,
      onClick: () => selectItem(id),
      href: "#" + id,
      ...data
    };
  };

  const sideNav = [
    createItem("Edit Configurations", "config", {
      onClick: undefined,
      icon: <EuiIcon type="indexEdit" />,
      items: [
        createItem("Pre-Processing", "preprocessing", {
          icon: <EuiIcon type="editorItemAlignRight" />,
          forceOpen: true,
          items: [
            createItem("Input", "input-preprocess", {
              icon: <EuiIcon type="logstashInput" />
            }),
            createItem("Transformation", "transform-preprocess", {
              icon: <EuiIcon type="inputOutput" />
            }),
            createItem("Output", "output-preprocess", {
              icon: <EuiIcon type="logstashOutput" />
            })
          ]
        }),
        createItem("Post-Processing", "postprocessing", {
          icon: <EuiIcon type="editorItemAlignLeft" />,
          forceOpen: true,
          items: [
            createItem("Input", "input-postprocess", {
              icon: <EuiIcon type="logstashInput" />
            }),
            createItem("Transformation", "transform-postprocess", {
              icon: <EuiIcon type="inputOutput" />
            }),
            createItem("Output", "output-postprocess", {
              icon: <EuiIcon type="logstashOutput" />
            })
          ]
        }),
        createItem("Simulation", "simulation", {
          icon: <EuiIcon type="play" />
        })
      ]
    })
  ];

  const createHeader = (title, subtitle, iconType) => {
    return (
      <div>
        <EuiFlexGroup gutterSize="s" alignItems="center" responsive={false}>
          <EuiFlexItem grow={false}>
            <EuiIcon type={iconType} size="m" />
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiTitle size="xs">
              <h3>{title}</h3>
            </EuiTitle>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiText size="s">
          <p>
            <EuiTextColor color="subdued">{subtitle}</EuiTextColor>
          </p>
        </EuiText>
      </div>
    );
  };

  return (
    <EuiFlexGroup>
      <EuiFlexItem>
        <div className="sidebar">
          <EuiSideNav
            aria-label="Standard Transformer Config"
            mobileTitle="Standard Transformer Config"
            toggleOpenOnMobile={() => toggleOpenOnMobile()}
            isOpenOnMobile={isSideNavOpenOnMobile}
            style={{ width: 192 }}
            items={sideNav}
          />
        </div>
      </EuiFlexItem>

      <EuiFlexItem grow={7}>
        <div className="config">
          <EuiFlexGroup direction="column" gutterSize="m">
            <div id="preprocessing" className="preprocessing">
              <EuiAccordion
                id="preprocess"
                element="fieldset"
                className="euiAccordionForm"
                buttonClassName="euiAccordionForm__button"
                buttonContent={createHeader(
                  "Pre-Processing",
                  "Input, Transformation and Output configurations for pre-processing",
                  "editorItemAlignRight"
                )}
                paddingSize="l"
                initialIsOpen={true}>
                <PipelineStage
                  stage="preprocess"
                  values={preValues}
                  onChangeHandler={onChange(
                    "transformer.config.transformerConfig.preprocess"
                  )}
                  errors={get(
                    errors,
                    "transformer.config.transformerConfig.preprocess"
                  )}
                />
              </EuiAccordion>
            </div>

            <div id="postprocessing" className="postprocessing">
              <EuiAccordion
                id="postprocess"
                element="fieldset"
                className="euiAccordionForm"
                buttonClassName="euiAccordionForm__button"
                buttonContent={createHeader(
                  "Post-Processing",
                  "Input, Transformation and Output configurations for post-processing",
                  "editorItemAlignLeft"
                )}
                paddingSize="l"
                initialIsOpen={true}>
                <PipelineStage
                  stage="postprocess"
                  values={postValues}
                  onChangeHandler={onChange(
                    "transformer.config.transformerConfig.postprocess"
                  )}
                  errors={get(
                    errors,
                    "transformer.config.transformerConfig.postprocess"
                  )}
                />
              </EuiAccordion>
            </div>

            <div id="simulation" className="simulation">
              <EuiAccordion
                id="simulation"
                element="fieldset"
                className="euiAccordionForm"
                buttonClassName="euiAccordionForm__button"
                buttonContent={createHeader(
                  "Simulation",
                  "Simulation of standard transformer given your config without deploying the model",
                  "play"
                )}
                paddingSize="l"
                initialIsOpen={true}>
                <TransformerSimulation />
              </EuiAccordion>
            </div>
          </EuiFlexGroup>
        </div>
      </EuiFlexItem>

      <EuiFlexItem grow={3}>
        <PipelineSidebarPanel />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
