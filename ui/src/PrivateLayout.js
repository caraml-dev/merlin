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
import { Outlet, useNavigate } from "react-router-dom";
import {
  ApplicationsContext,
  ApplicationsContextProvider,
  Header,
  PrivateRoute,
  ProjectsContextProvider
} from "@caraml-dev/ui-lib";
import urlJoin from "proper-url-join";
import { appConfig } from "./config";
import { EnvironmentsContextProvider } from "./providers/environments/context";

import "./PrivateLayout.scss";

export const PrivateLayout = () => {
  const navigate = useNavigate();
  return (
    <PrivateRoute>
      <ApplicationsContextProvider>
        <ProjectsContextProvider>
          <EnvironmentsContextProvider>
            <ApplicationsContext.Consumer>
              {({ currentApp }) => (
                <Header
                  appIcon="machineLearningApp"
                  docLinks={appConfig.docsUrl}
                  onProjectSelect={pId =>
                    navigate(urlJoin(currentApp?.homepage, "projects", pId, "models"))
                  }
              />)}
            </ApplicationsContext.Consumer>
            <Outlet />
          </EnvironmentsContextProvider>
        </ProjectsContextProvider>
      </ApplicationsContextProvider>
    </PrivateRoute>
  );
};
