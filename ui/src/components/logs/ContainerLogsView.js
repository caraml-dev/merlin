import React, {
  Fragment,
  useContext,
  useEffect,
  useMemo,
  useState
} from "react";
import {
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingContent,
  EuiPageContent,
  EuiSpacer,
  EuiTextColor,
  EuiTitle
} from "@elastic/eui";
import { LazyLog, ScrollFollow } from "react-lazylog";
import { AuthContext } from "@gojek/mlp-ui";
import config from "../../config";
import mocks from "../../mocks";
import { LogsSearchBar } from "./LogsSearchBar";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const querystring = require("querystring");

const componentOrder = [
  "image_builder",
  "model",
  "transformer",
  "batch_job_driver",
  "batch_job_executor"
];

export const ContainerLogsView = ({
  projectId,
  model,
  versionId,
  jobId,
  fetchContainerURL
}) => {
  const [{ data: project, isLoaded: projectLoaded }] = useMerlinApi(
    `/projects/${projectId}`,
    { mock: mocks.project },
    []
  );

  const [params, setParams] = useState({
    component_type: "",
    tail_lines: "1000"
  });

  const [{ data: containers, isLoaded: containersLoaded }] = useMerlinApi(
    fetchContainerURL,
    { mock: mocks.containerOptions },
    []
  );

  useEffect(
    () => {
      if (containersLoaded) {
        if (
          containers.find(
            container => container.component_type === "image_builder"
          )
        ) {
          setParams({ ...params, component_type: "image_builder" });
        }
        if (
          containers.find(container => container.component_type === "model")
        ) {
          setParams({ ...params, component_type: "model" });
        }
        if (
          containers.find(
            container => container.component_type === "batch_job_driver"
          )
        ) {
          setParams({ ...params, component_type: "batch_job_driver" });
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [containers, containersLoaded]
  );

  const componentTypes = useMemo(() => {
    return containers
      ? containers
          .sort(
            (a, b) =>
              componentOrder.indexOf(a.component_type) -
              componentOrder.indexOf(b.component_type)
          )
          .map(container => container.component_type)
      : [];
  }, [containers]);

  const authCtx = useContext(AuthContext);
  const fetchOptions = {
    headers: {
      Authorization: `Bearer ${authCtx.state.accessToken}`
    }
  };

  const [logUrl, setLogUrl] = useState("");
  useEffect(() => {
    if (params.component_type !== "" && projectLoaded) {
      const container = containers.find(
        container => container.component_type === params.component_type
      );
      if (container) {
        const containerQuery = {
          ...params,
          cluster: container.cluster,
          namespace: container.namespace,
          timestamps: true,
          model_id: model.id,
          model_name: model.name,
          version_id: versionId
        };
        const logParams = querystring.stringify(containerQuery);
        setLogUrl(config.MERLIN_API + "/logs?" + logParams);
      }
    }
  }, [params, containers, project, projectLoaded, model, versionId]);

  return (
    <Fragment>
      <EuiTitle size="s">
        <EuiTextColor color="secondary">Logs</EuiTextColor>
      </EuiTitle>

      <EuiPageContent>
        {!containersLoaded ? (
          <EuiLoadingContent lines={4} />
        ) : containers && containers.length > 0 ? (
          <EuiFlexGroup direction="column" gutterSize="none">
            <EuiFlexItem grow={false}>
              <LogsSearchBar {...{ componentTypes, params, setParams }} />
            </EuiFlexItem>

            <EuiFlexItem grow={false}>
              <EuiSpacer size="s" />
            </EuiFlexItem>

            <EuiFlexItem grow={true}>
              {logUrl && (
                <ScrollFollow
                  startFollowing={true}
                  render={({ onScroll, follow }) => (
                    <LazyLog
                      caseInsensitive
                      enableSearch
                      extraLines={1}
                      fetchOptions={fetchOptions}
                      follow={follow}
                      height={640}
                      onScroll={onScroll}
                      selectableLines
                      stream
                      url={logUrl}
                    />
                  )}
                />
              )}
            </EuiFlexItem>
          </EuiFlexGroup>
        ) : (
          <EuiEmptyPrompt
            title={<h2>You have no logs</h2>}
            body={
              <Fragment>
                <p>
                  We cannot find any logs because there is no active component
                  for your model deployment.
                </p>
              </Fragment>
            }
          />
        )}
      </EuiPageContent>
    </Fragment>
  );
};
