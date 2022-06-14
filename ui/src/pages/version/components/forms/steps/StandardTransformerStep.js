import {
  EuiAccordion,
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiIcon,
  EuiPanel,
  EuiSideNav,
  EuiSpacer,
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
import { Element, scroller } from "react-scroll";
import { simulationIcon } from "../components/transformer/components/simulation/constants";
import { TransformationGraph } from "../components/transformer/components/TransformationGraph";
import { TransformationSpec } from "../components/transformer/components/TransformationSpec";
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

  const [togglePanelButton, setTogglePanelButton] = useState({
    "transformer-sidebar-panel": "show",
    "transformer-graph-panel": "hide"
  });
  const onPanelChange = (panelId, state) => {
    const newtogglePanelButton = {
      ...togglePanelButton,
      [panelId]: state
    };
    setTogglePanelButton(newtogglePanelButton);
  };

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
      onClick: () => {
        selectItem(id);
        scroller.scrollTo(id, {
          offset: -75
        });
      },
      href: "#" + id,
      ...data
    };
  };

  const sideNav = [
    createItem("Edit Configurations", "config", {
      renderItem: () => (
        <EuiFlexGroup justifyContent="spaceBetween" alignItems="center">
          <EuiFlexItem>
            <EuiText size="xs">
              <strong>Transformer Configuration</strong>
            </EuiText>
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiButtonIcon
              color="text"
              aria-label={"Toggle sidebar panel"}
              iconType="menuLeft"
              onClick={() => onPanelChange("transformer-sidebar-panel", "hide")}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      ),
      items: [
        createItem("Pre-Processing", "preprocessing", {
          icon: <EuiIcon type="editorItemAlignRight" />,
          forceOpen: true,

          items: [
            createItem("Input", "input-preprocess", {
              icon: <EuiIcon type="logstashInput" />
            }),
            createItem("Transformation", "transform-preprocess", {
              icon: <EuiIcon type="tableDensityExpanded" />
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
              icon: <EuiIcon type="tableDensityExpanded" />
            }),
            createItem("Output", "output-postprocess", {
              icon: <EuiIcon type="logstashOutput" />
            })
          ]
        }),
        createItem("Simulation", "simulation", {
          icon: <EuiIcon type={simulationIcon} />
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
      {togglePanelButton &&
      togglePanelButton["transformer-sidebar-panel"] === "hide" ? (
        <EuiButtonIcon
          color="text"
          aria-label={"Toggle sidebar panel"}
          iconType="menuRight"
          onClick={() => onPanelChange("transformer-sidebar-panel")}
          className="toggle-panel-button"
        />
      ) : (
        <EuiFlexItem grow={1}>
          <div className="config-sidebar">
            <EuiPanel grow={false} hasShadow={false}>
              <EuiSideNav
                aria-label="Standard Transformer Config"
                mobileTitle="Standard Transformer Config"
                toggleOpenOnMobile={() => toggleOpenOnMobile()}
                isOpenOnMobile={isSideNavOpenOnMobile}
                items={sideNav}
              />
            </EuiPanel>

            <EuiSpacer />

            <EuiPanel grow={false} hasShadow={false}>
              <TransformationSpec importEnabled={true} />
            </EuiPanel>
          </div>
        </EuiFlexItem>
      )}

      <EuiFlexItem grow={7}>
        <EuiPanel grow={false} paddingSize="none" hasShadow={false}>
          <div className="config">
            <Element name="preprocessing" className="preprocessing">
              <EuiAccordion
                id="preprocess-accordion"
                element="fieldset"
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
            </Element>

            <Element name="postprocessing" className="postprocessing">
              <EuiAccordion
                id="postprocess-accordion"
                element="fieldset"
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
            </Element>

            <Element name="simulation" className="simulation">
              <EuiAccordion
                id="simulation-accordion"
                element="fieldset"
                buttonClassName="euiAccordionForm__button"
                buttonContent={createHeader(
                  "Simulation",
                  "Simulation of standard transformer given your config without deploying the model",
                  simulationIcon
                )}
                paddingSize="l"
                initialIsOpen={true}>
                <TransformerSimulation />
              </EuiAccordion>
            </Element>
          </div>
        </EuiPanel>
      </EuiFlexItem>

      {togglePanelButton &&
      togglePanelButton["transformer-graph-panel"] === "hide" ? (
        <EuiButtonIcon
          color="text"
          aria-label={"Toggle graph panel"}
          iconType="menuLeft"
          onClick={() => onPanelChange("transformer-graph-panel")}
          className="toggle-panel-button"
        />
      ) : (
        <EuiFlexItem grow={2}>
          <div className="graph-sidebar">
            <EuiPanel grow={false} hasShadow={false}>
              <EuiFlexGroup justifyContent="spaceBetween" alignItems="center">
                <EuiFlexItem>
                  <EuiText size="xs">
                    <strong>Transformer Graph</strong>
                  </EuiText>
                </EuiFlexItem>
                <EuiFlexItem grow={false}>
                  <EuiButtonIcon
                    color="text"
                    aria-label={"Toggle graph panel"}
                    iconType="menuRight"
                    onClick={() =>
                      onPanelChange("transformer-graph-panel", "hide")
                    }
                  />
                </EuiFlexItem>
              </EuiFlexGroup>
              <TransformationGraph />
            </EuiPanel>
          </div>
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
