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
import StackdriverLink from "./StackdriverLink";
import { createStackdriverUrl } from "../../utils/createStackdriverUrl";

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

  const [{ data: newContainers }, getNewContainers] = useMerlinApi(
    fetchContainerURL,
    { mock: mocks.containerOptions },
    []
  );

  useEffect(() => {
    var handle = setInterval(getNewContainers, 5000);
    return () => {
      clearInterval(handle);
    };
  }, [getNewContainers]);

  const componentTypes = useMemo(() => {
    return newContainers
      ? newContainers
          .sort(
            (a, b) =>
              componentOrder.indexOf(a.component_type) -
              componentOrder.indexOf(b.component_type)
          )
          .map(container => container.component_type)
      : [];
  }, [newContainers]);

  const authCtx = useContext(AuthContext);
  const fetchOptions = {
    headers: {
      Authorization: `Bearer ${authCtx.state.accessToken}`
    }
  };

  const [logUrl, setLogUrl] = useState("");
  const [stackdriverUrl, setStackdriverUrl] = useState("");

  useEffect(() => {
    if (params.component_type !== "" && projectLoaded) {
      const activeContainers = containers.filter(
        container => container.component_type === params.component_type
      );

      if (activeContainers && activeContainers.length > 0) {
        const containerQuery = {
          ...params,
          cluster: activeContainers[0].cluster,
          namespace: activeContainers[0].namespace,
          timestamps: true,
          project_name: project.name,
          model_id: model.id,
          model_name: model.name,
          version_id: versionId,
          prediction_job_id: jobId
        };
        const logParams = querystring.stringify(containerQuery);
        setLogUrl(config.MERLIN_API + "/logs?" + logParams);

        const pods = [
          ...new Set(
            activeContainers.map(container => `"${container.pod_name}"`)
          )
        ];
        let stackdriverQuery = {
          gcp_project: activeContainers[0].gcp_project,
          cluster: activeContainers[0].cluster,
          namespace: activeContainers[0].namespace,
          pod_name: pods.join(" OR ")
        };
        setStackdriverUrl(createStackdriverUrl(stackdriverQuery));
      }
    }
  }, [params, containers, project, projectLoaded, model, versionId, jobId]);

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

            {logUrl && (
              <EuiFlexItem grow={true}>
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
              </EuiFlexItem>
            )}

            {stackdriverUrl && (
              <EuiFlexItem grow={false}>
                <StackdriverLink stackdriverUrl={stackdriverUrl} />
              </EuiFlexItem>
            )}
          </EuiFlexGroup>
        ) : (
          <EuiEmptyPrompt
            title={<h2>You have no logs</h2>}
            body={
              <Fragment>
                <p>
                  We cannot find any logs because there is no active component
                  for your {jobId ? "batch job" : "model deployment"}.
                </p>
              </Fragment>
            }
          />
        )}
      </EuiPageContent>
    </Fragment>
  );
};
