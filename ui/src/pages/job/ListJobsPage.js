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

import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import {
  EuiButton,
  EuiPageTemplate,
  EuiPanel,
  EuiSearchBar,
  EuiSpacer,
} from "@elastic/eui";
import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { appConfig } from "../../config";
import { useMerlinApi } from "../../hooks/useMerlinApi";
import mocks from "../../mocks";
import JobSearchBar from "./components/JobSearchBar";
import JobsTable from "./components/JobsTable";
import { generateBreadcrumbs } from "./utils/breadcrumbs";

const ListJobPage = () => {
  const { projectId, modelId } = useParams();
  const createJobURL = `/merlin/projects/${projectId}/models/${modelId}/create-job`;

  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  /* To capture the search text from the search bar so that the search can be
     done on the backend. */
  const [searchText, setSearchText] = useState("");
  const [searchStatus, setSearchStatus] = useState("");

  const onSearchTextChange = (query) => {
    setPage({ ...page, index: 0 });

    const queryString = EuiSearchBar.Query.toESQueryString(query);
    if (queryString === "*") {
      setSearchText("");
      setSearchStatus("");
      return;
    }

    const params = queryString.split("+");
    params.forEach((param) => {
      const p = param.trim();
      if (p.includes("status:")) {
        setSearchStatus(p.split("status:")[1]);
      } else if (p !== "") {
        setSearchText(p);
      }
    });
  };

  const [{ data, isLoaded, error }, fetchJobs] = useMerlinApi(
    `/projects/${projectId}/jobs-by-page?model_id=${modelId}&page=${page.index + 1}&page_size=${page.size}&search=${searchText}&status=${searchStatus}`,
    { mock: mocks.jobList },
    []
  );

  const [{ data: model }] = useMerlinApi(
    `/projects/${projectId}/models/${modelId}`,
    { mock: mocks.model },
    []
  );

  useEffect(() => {
    replaceBreadcrumbs(
      generateBreadcrumbs(projectId, modelId, model, null, null)
    );
  }, [projectId, modelId, model]);

  return (
    <EuiPageTemplate restrictWidth="90%" paddingSize="none">
      <EuiSpacer size="l" />

      <EuiPageTemplate.Header
        bottomBorder={false}
        iconType={"machineLearningApp"}
        pageTitle="Prediction Jobs"
        rightSideItems={[
          <EuiButton fill href={createJobURL}>
            Start Batch Job
          </EuiButton>,
        ]}
      />

      <EuiSpacer size="l" />

      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <JobSearchBar
            placeholder="Search batch prediction job"
            onChange={({ query, queryText }) => {
              onSearchTextChange(query, queryText);
            }}
          />

          <EuiSpacer size="s" />

          <JobsTable
            projectId={projectId}
            modelId={modelId}
            jobs={data.results || []}
            isLoaded={isLoaded}
            error={error}
            page={page}
            totalItemCount={data?.paging?.total || 0}
            onPaginationChange={setPage}
            searchText={searchText}
            onSearchTextChange={onSearchTextChange}
            fetchJobs={fetchJobs}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>

      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default ListJobPage;
