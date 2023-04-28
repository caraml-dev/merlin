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

import { EuiProvider } from "@elastic/eui";
import {
  AuthProvider,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  Page404,
  Toast,
} from "@caraml-dev/ui-lib";
import { Route, Routes } from "react-router-dom";
import React from "react";
import config from "./config";
import { PrivateLayout } from "./PrivateLayout";
import AppRoutes from "./AppRoutes";

const App = () => {
  return (
    <EuiProvider>
      <ErrorBoundary>
        <MlpApiContextProvider
          mlpApiUrl={config.MLP_API}
          timeout={config.TIMEOUT}
          useMockData={config.USE_MOCK_DATA}>
          <AuthProvider clientId={config.OAUTH_CLIENT_ID}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route element={<PrivateLayout />}>
                <Route path="/*" element={<AppRoutes />} />
              </Route>
              <Route path="/pages/404" element={<Page404 />} />
            </Routes>
            <Toast />
          </AuthProvider>
        </MlpApiContextProvider>
      </ErrorBoundary>
    </EuiProvider>
  );
};

export default App;
