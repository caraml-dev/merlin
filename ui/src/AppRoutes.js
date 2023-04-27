import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import config from "./config";
import Home from "./Home";
import Models from "./model/Models";
import { ModelDetails } from "./model/ModelDetails";
import Versions from "./version/Versions";
import { CreateJobView } from "./job/CreateJobView";
import JobDetails from "./job/JobDetails";
import Jobs from "./job/Jobs";

// The new UI architecture will have all UI pages inside of `pages` folder
import {
  DeployModelVersionView,
  RedeployModelVersionView,
  TransformerTools,
  VersionDetails,
} from "./pages";

const AppRoutes = () => {
  return (
    <Routes>
      <Route path={config.HOMEPAGE}>
        <Route index element={<Home />} />
        <Route path="projects/:projectId">
          <Route index element={<Navigate to="models" replace={true} />} />
          {/* TRANSFORMER SIMULATOR */}
          <Route path="transformer-simulator" element={<TransformerTools />} />
          {/* MODELS */}
          <Route path="models">
            {/* LIST */}
            <Route index element={<Models />} />
            <Route path=":modelId/*" element={<ModelDetails />} />
            {/* VERSIONS */}
            <Route path=":modelId/versions/*" element={<Versions />} />
            <Route path=":modelId/versions/:versionId/*" element={<VersionDetails />} />
            <Route path=":modelId/versions/:versionId/deploy" element={<DeployModelVersionView />} />
            {/* VERSIONS ENDPOINTS */}
            <Route path=":modelId/versions/:versionId/endpoints">
              <Route index={true} path=":endpointId/*" element={<VersionDetails />} />
              <Route path=":endpointId/redeploy" element={<RedeployModelVersionView />} />
            </Route>
            {/* BATCH JOBS */}
            <Route path=":modelId/versions/:versionId/jobs" element={<Jobs />} />
            <Route path=":modelId/versions/:versionId/jobs/:jobId/*" element={<JobDetails />} />
            <Route path=":modelId/create-job" element={<CreateJobView />} />
            <Route path=":modelId/versions/:versionId/create-job" element={<CreateJobView />} />
            {/* REDIRECTS */}
            <Route path=":modelId" element={<Navigate to="versions" replace={true} />} />
            <Route path=":modelId/versions/:versionId" element={<Navigate to="details" replace={true} />} />
            <Route path=":modelId/versions/:versionId/endpoints/:endpointId" element={<Navigate to="details" replace={true} />} />
          </Route>
        </Route>
      </Route>
      {/* For backward compatibility */}
      <Route path="-/tools/transformer" element={<TransformerTools />} />
      {/* DEFAULT */}
      <Route path="*" element={<Navigate to="/pages/404" replace={true} />} />
    </Routes>
  );
};

export default AppRoutes;
