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

import React from "react";
import { Redirect, Router } from "@reach/router";
import {
  AuthProvider,
  Page404,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  PrivateRoute,
  Toast
} from "@gojek/mlp-ui";
import config from "./config";
import Home from "./Home";
import Models from "./model/Models";
import { ModelDetails } from "./model/ModelDetails";
import Versions from "./version/Versions";
import Jobs from "./job/Jobs";
import JobDetails from "./job/JobDetails";
import { CreateJobView } from "./job/CreateJobView";
import { PrivateLayout } from "./PrivateLayout";
import { EuiProvider } from "@elastic/eui";

// The new UI architecture will have all UI pages inside of `pages` folder
import {
  DeployModelVersionView,
  RedeployModelVersionView,
  VersionDetails,
  TransformerTools
} from "./pages";

const App = () => {

  return(
    <EuiProvider>
      <ErrorBoundary>
        <MlpApiContextProvider
          mlpApiUrl={config.MLP_API}
          timeout={config.TIMEOUT}
          useMockData={config.USE_MOCK_DATA}>
          <AuthProvider clientId={config.OAUTH_CLIENT_ID}>
            <Router role="group">
              <Login path="/login" />

              <Redirect from="/" to={config.HOMEPAGE} noThrow />

              <PrivateRoute path={config.HOMEPAGE} render={PrivateLayout(Home)} />

              <Redirect
                from={`${config.HOMEPAGE}/projects/:projectId`}
                to={`${config.HOMEPAGE}/projects/:projectId/models`}
                noThrow
              />

              {/* MODELS */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models`}
                render={PrivateLayout(Models)}
              />

              {/* MODELS DETAILS (Sub-router) */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/*`}
                render={PrivateLayout(ModelDetails)}
              />
              <Redirect
                from={`${config.HOMEPAGE}/projects/:projectId/models/:modelId`}
                to={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions`}
                noThrow
              />

              {/* VERSIONS */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions`}
                render={PrivateLayout(Versions)}
              />

              {/* Deploy model version */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/deploy`}
                render={PrivateLayout(DeployModelVersionView)}
              />

              {/* Redeploy model version */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/endpoints/:endpointId/redeploy`}
                render={PrivateLayout(RedeployModelVersionView)}
              />

              {/* Version pages (and its sub-routers) */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/*`}
                render={PrivateLayout(VersionDetails)}
              />

              <Redirect
                from={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId`}
                to={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/details`}
                noThrow
              />

              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/endpoints/:endpointId/*`}
                render={PrivateLayout(VersionDetails)}
              />

              <Redirect
                from={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/endpoints/:endpointId`}
                to={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/endpoints/:endpointId/details`}
                noThrow
              />

              {/* Prediction Job */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/jobs`}
                render={PrivateLayout(Jobs)}
              />

              {/* Prediction Job Details (Sub-router) */}
              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/jobs/:jobId/*`}
                render={PrivateLayout(JobDetails)}
              />

              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/create-job`}
                render={PrivateLayout(CreateJobView)}
              />

              <PrivateRoute
                path={`${config.HOMEPAGE}/projects/:projectId/models/:modelId/versions/:versionId/create-job`}
                render={PrivateLayout(CreateJobView)}
              />

              {/* Tools */}
              <TransformerTools path={`${config.HOMEPAGE}/-/tools/transformer`} />

              {/* DEFAULT */}
              <Page404 default />
            </Router>
            <Toast />
          </AuthProvider>
        </MlpApiContextProvider>
      </ErrorBoundary>
    </EuiProvider>
  );
};

export default App;
