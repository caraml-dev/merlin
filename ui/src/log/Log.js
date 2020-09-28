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

import React, { Fragment, useEffect, useState } from "react";
import {
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFlyout,
  EuiFlyoutBody,
  EuiFlyoutFooter,
  EuiFlyoutHeader,
  EuiFormRow,
  EuiIcon,
  EuiLink,
  EuiSuperSelect,
  EuiTitle,
  EuiText,
  EuiPageContent
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import config from "../config";
import mocks from "../mocks";
import { createStackdriverUrl } from "../utils/createStackdriverUrl";
import StackdriverLink from "./components/StackdriverLink";
import StreamLog from "./components/StreamLog";
import { useMerlinApi } from "../hooks/useMerlinApi";

const querystring = require("querystring");

const NO_CONTAINER = "no-container";

const Log = ({ modelId, fetchContainerURL, breadcrumbs }) => {
  const [{ data: containers }, fetchContainers] = useMerlinApi(
    fetchContainerURL,
    { mock: mocks.containerOptions },
    []
  );

  // Container options for SuperSelect UI component
  // Default value is "No Container"
  const [containerOptions, setContainerOptions] = useState([
    {
      value: NO_CONTAINER,
      inputDisplay:
        "No container. The deployment is scaled down to zero due to inactivity.",
      dropdownDisplay: (
        <strong>
          No container. The deployment is scaled down to zero due to inactivity.
        </strong>
      ),
      disabled: true
    }
  ]);
  const [selectedContainer, setSelectedContainer] = useState(NO_CONTAINER);
  const [logUrl, setLogUrl] = useState("");
  const [stackdriverUrl, setStackdriverUrl] = useState("");

  // Upon receiving containers response from Merlin API, update container options to be displayed.
  useEffect(() => {
    if (containers && containers.length > 0) {
      setContainerOptions(() => {
        return containers
          .sort((a, b) => (a.pod_name < b.pod_name ? 1 : -1))
          .sort((a, b) => (a.name > b.name ? 1 : -1))
          .sort(a =>
            a.name === "kfserving-container" ||
            a.name === "user-container" ||
            a.name === "spark-kubernetes-driver"
              ? -1
              : 1
          )
          .map(container => {
            return {
              value: JSON.stringify(container),
              inputDisplay: `${container.name} (${container.pod_name})`,
              dropdownDisplay: (
                <Fragment>
                  <strong>{container.name}</strong>
                  <EuiText size="xs">
                    <p className="euiTextColor--subdued">
                      Pod name: {container.pod_name}
                    </p>
                  </EuiText>
                </Fragment>
              )
            };
          });
      });
    }
  }, [containers]);

  useEffect(() => {
    replaceBreadcrumbs([...breadcrumbs, { text: "Logs" }]);
  }, [breadcrumbs]);

  const onSelectContainerChange = value => {
    setSelectedContainer(value);

    if (value !== NO_CONTAINER) {
      const containerQuery = JSON.parse(value);
      containerQuery.follow = true;
      containerQuery.timestamps = true;
      containerQuery.model_id = modelId;
      const logParams = querystring.stringify(containerQuery);
      setLogUrl(config.MERLIN_API + "/logs?" + logParams);

      setStackdriverUrl(createStackdriverUrl(containerQuery));
    }
  };

  useEffect(() => {
    if (
      containerOptions &&
      containerOptions.length > 0 &&
      selectedContainer === NO_CONTAINER
    ) {
      onSelectContainerChange(containerOptions[0].value);
    }
    // eslint-disable-next-line
  }, [containerOptions]);

  const [isFlyoutVisible, setIsFlyoutVisible] = useState(false);

  let flyout = (
    <EuiFlyout
      onClose={() => setIsFlyoutVisible(false)}
      aria-labelledby="model-container-help"
      size="m"
      ownFocus>
      <EuiFlyoutHeader hasBorder>
        <EuiTitle size="s">
          <h2 id="model-container-help">Model containers</h2>
        </EuiTitle>
      </EuiFlyoutHeader>

      <EuiFlyoutBody>
        <EuiText size="s">
          <p>
            Machine Learning Model is deployed as container within a{" "}
            <a
              href="https://kubernetes.io/docs/concepts/workloads/pods/pod-overview/"
              target="_blank"
              rel="noopener noreferrer">
              Kubernetes Pod
            </a>
            .{" "}
            <a
              href="https://kubernetes.io/docs/concepts/containers/overview/"
              target="_blank"
              rel="noopener noreferrer">
              Containers
            </a>{" "}
            are a technology for packaging the (compiled) code for an
            application along with the dependencies it needs.
          </p>

          <p>
            This is the list of containers that Merlin needs to make sure your
            ML model running as expected:
          </p>

          <ol>
            <li>
              <b>kfserving-container</b>: The main container where your model is
              running.
            </li>
            <li>
              <b>user-container</b>: For Pyfunc model only. This container is
              similar as kfserving-container but with different name.
            </li>
            <li>
              <b>queue-proxy</b>: As an HTTP request proxy to ML model.
            </li>
            <li>
              <b>storage-initializer</b>: To download the models and artifacts
              needed to run your model.
            </li>
            <li>
              <b>pyfunc-image-builder</b>: As the name suggests, this container
              builds the container image for Pyfunc ML model. This container
              only runs once at the beginning of Pyfunc model deployment.
            </li>
            <li>
              <b>spark-kubernetes-driver</b>: Schedules the task of batch
              prediction job to be executed by executor.
            </li>
            <li>
              <b>executor</b>: This container executes the batch prediction job.
            </li>
          </ol>

          <p>
            You can see the logs for each active container. If there's no
            container available, it might be because your model is inactive and
            scaled to zero. You can activate your model by sending an HTTP
            request to its endpoint.
          </p>
        </EuiText>
      </EuiFlyoutBody>

      <EuiFlyoutFooter>
        <EuiFlexGroup justifyContent="spaceBetween">
          <EuiFlexItem grow={false}>
            <EuiButtonEmpty
              iconType="cross"
              onClick={() => setIsFlyoutVisible(false)}
              flush="left">
              Close
            </EuiButtonEmpty>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlyoutFooter>
    </EuiFlyout>
  );

  return (
    <EuiPageContent>
      <EuiFlexGroup direction="column">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <EuiText size="xs">
                Select container{" "}
                <EuiLink
                  color="secondary"
                  onClick={() => setIsFlyoutVisible(true)}>
                  <EuiIcon type="questionInCircle" color="subdued" />
                </EuiLink>
              </EuiText>
            }>
            <EuiSuperSelect
              options={containerOptions}
              valueOfSelected={selectedContainer}
              onChange={onSelectContainerChange}
              onFocus={() => {
                fetchContainers();
              }}
              hasDividers
              fullWidth
            />
          </EuiFormRow>
        </EuiFlexItem>

        {stackdriverUrl !== "" && (
          <EuiFlexItem>
            <StackdriverLink stackdriverUrl={stackdriverUrl} />
          </EuiFlexItem>
        )}

        {logUrl !== "" && (
          <EuiFlexItem>
            <StreamLog logUrl={logUrl} />
          </EuiFlexItem>
        )}

        {isFlyoutVisible && flyout}
      </EuiFlexGroup>
    </EuiPageContent>
  );
};

export default Log;
