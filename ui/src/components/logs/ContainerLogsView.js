import { AuthContext } from "@caraml-dev/ui-lib";
import {
  EuiEmptyPrompt,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiLoadingContent,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import React, { Fragment, useContext, useEffect, useState } from "react";
import { LazyLog, ScrollFollow } from "react-lazylog";
import config from "../../config";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";
import { createStackdriverUrl } from "../../utils/createStackdriverUrl";
import { LogsSearchBar } from "./LogsSearchBar";

const componentOrder = [
  "image_builder",
  "model",
  "transformer",
  "batch_job_driver",
  "batch_job_executor",
];

export const ContainerLogsView = ({
  projectId,
  model,
  versionId,
  revisionId,
  jobId,
  fetchContainerURL,
}) => {
  const [{ data: project, isLoaded: projectLoaded }] = useMerlinApi(
    `/projects/${projectId}`,
    { mock: mocks.project },
    []
  );

  const [params, setParams] = useState({
    component_type: "",
    tail_lines: "1000",
  });

  const [containerHaveBeenLoaded, setContainerHaveBeenLoaded] = useState(false);

  const [{ data: containers, isLoaded: containersLoaded }, getContainers] =
    useMerlinApi(fetchContainerURL, { mock: mocks.containerOptions }, [], true);

  useEffect(() => {
    var handle = setInterval(getContainers, 5000);
    return () => {
      clearInterval(handle);
    };
  }, [getContainers]);

  useEffect(
    () => {
      if (!containerHaveBeenLoaded) {
        setContainerHaveBeenLoaded(containersLoaded);
      }
      if (containersLoaded && params.component_type === "") {
        if (
          containers.find(
            (container) => container.component_type === "image_builder"
          )
        ) {
          setParams({ ...params, component_type: "image_builder" });
        }
        if (
          containers.find((container) => container.component_type === "model")
        ) {
          setParams({ ...params, component_type: "model" });
        } else if (
          containers.find(
            (container) => container.component_type === "transformer"
          )
        ) {
          setParams({ ...params, component_type: "transformer" });
        }
        if (
          containers.find(
            (container) => container.component_type === "batch_job_driver"
          )
        ) {
          setParams({ ...params, component_type: "batch_job_driver" });
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [containers, containersLoaded]
  );

  const [componentTypes, setComponentTypes] = useState();
  useEffect(() => {
    if (containers) {
      const newComponentTypes = containers
        .sort(
          (a, b) =>
            componentOrder.indexOf(a.component_type) -
            componentOrder.indexOf(b.component_type)
        )
        .map((container) => container.component_type);
      if (
        JSON.stringify(newComponentTypes) !== JSON.stringify(componentTypes)
      ) {
        setComponentTypes(newComponentTypes);
      }
    }
  }, [containers, componentTypes]);

  const authCtx = useContext(AuthContext);
  const fetchOptions = {
    headers: {
      Authorization: `Bearer ${authCtx.state.jwt}`,
    },
  };

  const [logUrl, setLogUrl] = useState("");
  const [stackdriverUrls, setStackdriverUrls] = useState({});
  useEffect(
    () => {
      if (projectLoaded) {
        // set image builder url
        let stackdriverQuery = {
          job_name: project.name + "-" + model.name + "-" + versionId,
          start_time: model.updated_at,
        };
        setStackdriverUrls({...stackdriverUrls, "image_builder":createStackdriverUrl(stackdriverQuery, "image_builder")});

        // update active container
        if (params.component_type !== "") {
          const activeContainers = containers.filter(
            (container) => container.component_type === params.component_type
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
              revision_id: revisionId ? revisionId : "",
              prediction_job_id: jobId ? jobId : "",
            };
            const logParams = new URLSearchParams(containerQuery).toString();
            const newLogUrl = config.MERLIN_API + "/logs?" + logParams;
            if (newLogUrl !== logUrl) {
              setLogUrl(newLogUrl);
            }

            const pods = [
              ...new Set(
                activeContainers.map((container) => `"${container.pod_name}"`)
              ),
            ];
            let stackdriverQuery = {
              gcp_project: activeContainers[0].gcp_project,
              cluster: activeContainers[0].cluster,
              namespace: activeContainers[0].namespace,
              pod_name: pods.join(" OR "),
              start_time: model.updated_at,
            };
            if (params.component_type !== "image_builder"){
              setStackdriverUrls({...stackdriverUrls, [params.component_type]: createStackdriverUrl(stackdriverQuery,params.component_type)});
            }
          }
        }
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [params, containers, project, projectLoaded, model, versionId, jobId]
  );

  return (
    <Fragment>
      <EuiTitle size="s">
        <span>
          <EuiTextColor color="success">&nbsp; Logs</EuiTextColor>
        </span>
      </EuiTitle>
      {
        Object.keys(stackdriverUrls).length !== 0 &&
        (
          <Fragment>
            <EuiSpacer size="s"/>
            <EuiPanel>
              <EuiFlexGroup direction="row" alignItems="center">
                <EuiFlexItem style={{marginTop:0, marginBottom:0}} grow={false}>
                  <EuiText  style={{ fontSize: '14px', fontWeight:"bold"}}>Stackdriver Logs</EuiText>
                </EuiFlexItem>
                {Object.entries(stackdriverUrls).map(([component,url])=> (
                  <EuiFlexItem style={{marginTop:0, marginBottom:0, paddingLeft:"10px", textTransform: "capitalize"}} key={component} grow={false}>
                    <EuiText size="xs" >
                      <EuiLink href={url} target="_blank" external>{component.replace(new RegExp("_", "g"), " ")}</EuiLink>
                    </EuiText>
                  </EuiFlexItem>
                ))}
              </EuiFlexGroup>
            </EuiPanel>
          </Fragment>
        )
      }
      <EuiSpacer size="s" />
      <EuiPanel>
        {!containerHaveBeenLoaded &&
        componentTypes &&
        componentTypes.length === 0 ? (
          <EuiLoadingContent lines={4} />
        ) : logUrl ? (
          <EuiFlexGroup direction="column" gutterSize="none">
            <EuiFlexItem grow={false}>
              <LogsSearchBar {...{ componentTypes, params, setParams }} />
            </EuiFlexItem>

            <EuiFlexItem grow={false}>
              <EuiSpacer size="s" />
            </EuiFlexItem>

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
          </EuiFlexGroup>
        ) : (
          <EuiEmptyPrompt
            title={<h2>Active Container Logs</h2>}
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
      </EuiPanel>
    </Fragment>
  );
};
