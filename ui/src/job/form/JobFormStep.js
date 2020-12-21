/**
 * Copyright 2020 The Merlin Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { Fragment } from "react";
import {
  EuiButton,
  EuiFlexGroup,
  EuiFlexItem,
  EuiHorizontalRule,
  EuiSpacer
} from "@elastic/eui";
import PropTypes from "prop-types";

export const JobFormStep = ({
  step,
  currentStep,
  isLastStep,
  onPrev,
  onNext,
  onSubmit,
  isStepCompleted,
  ...props
}) => {
  return currentStep === step ? (
    <Fragment>
      {props.children}

      <EuiHorizontalRule margin="s" />
      <EuiSpacer size="m" />

      <EuiFlexGroup direction="row" justifyContent="flexEnd">
        {step > 1 && (
          <EuiFlexItem grow={false}>
            <EuiButton
              size="s"
              color="primary"
              disabled={false}
              onClick={onPrev}>
              Previous
            </EuiButton>
          </EuiFlexItem>
        )}

        <EuiFlexItem grow={false}>
          {isLastStep ? (
            <EuiButton
              size="s"
              color="primary"
              fill
              disabled={!isStepCompleted}
              onClick={onSubmit}>
              Submit
            </EuiButton>
          ) : (
            <EuiButton
              size="s"
              color="primary"
              fill
              disabled={!isStepCompleted}
              onClick={onNext}>
              Next
            </EuiButton>
          )}
        </EuiFlexItem>
      </EuiFlexGroup>
    </Fragment>
  ) : null;
};

JobFormStep.propTypes = {
  step: PropTypes.number,
  currentStep: PropTypes.number,
  isLastStep: PropTypes.bool,
  onPrev: PropTypes.func,
  onNext: PropTypes.func,
  onSubmit: PropTypes.func,
  isStepCompleted: PropTypes.bool
};
