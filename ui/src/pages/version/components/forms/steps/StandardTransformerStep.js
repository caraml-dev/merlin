import React, { useContext, useState } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiSideNav,
  slugify,
  EuiIcon,
  EuiAccordion,
  EuiTitle
} from "@elastic/eui";
import { EuiText, EuiTextColor } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@gojek/mlp-ui";
import { PipelineSidebarPanel } from "../components/transformer/PipelineSidebarPanel";
import { PipelineStage } from "./PipelineStage";
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

  const createItem = (name, data = {}) => {
    // NOTE: Duplicate `name` values will cause `id` collisions.
    return {
      id: slugify(name),
      name,
      isSelected: selectedItemName === name,
      onClick: () => selectItem(name),
      href: "#" + name,
      ...data
    };
  };

  const sideNav = [
    createItem("Edit Configurations", {
      onClick: undefined,
      icon: <EuiIcon type="indexEdit" />,
      items: [createItem("Pre-Processing"), createItem("Post-Processing")]
    })
  ];

  const preprocessHeader = (
    <div>
      <EuiFlexGroup gutterSize="s" alignItems="center" responsive={false}>
        <EuiFlexItem grow={false}>
          <EuiIcon type="menuLeft" size="m" />
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiTitle size="xs">
            <h3>Pre-Processing</h3>
          </EuiTitle>
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiText size="s">
        <p>
          <EuiTextColor color="subdued">
            Input, Transformation and Output configurations for pre-processing
          </EuiTextColor>
        </p>
      </EuiText>
    </div>
  );

  const postprocessHeader = (
    <div>
      <EuiFlexGroup gutterSize="s" alignItems="center" responsive={false}>
        <EuiFlexItem grow={false}>
          <EuiIcon type="menuRight" size="m" />
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiTitle size="xs">
            <h3>Post-Processing</h3>
          </EuiTitle>
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiText size="s">
        <p>
          <EuiTextColor color="subdued">
            Input, Transformation and Output configurations for post-processing
          </EuiTextColor>
        </p>
      </EuiText>
    </div>
  );

  return (
    <EuiFlexGroup>
      <EuiFlexItem>
        <div class="sidebar">
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
        <div class="config">
          <EuiFlexGroup direction="column" gutterSize="m">
            <div id="Pre-Processing">
              <EuiAccordion
                id="preprocess"
                element="fieldset"
                className="euiAccordionForm"
                buttonClassName="euiAccordionForm__button"
                buttonContent={preprocessHeader}
                paddingSize="l"
                initialIsOpen="true">
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
            <div id="Post-Processing">
              <EuiAccordion
                id="postprocess"
                element="fieldset"
                className="euiAccordionForm"
                buttonClassName="euiAccordionForm__button"
                buttonContent={postprocessHeader}
                paddingSize="l"
                initialIsOpen="true">
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
          </EuiFlexGroup>
        </div>
      </EuiFlexItem>

      <EuiFlexItem grow={3}>
        <PipelineSidebarPanel />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
