import React, { useEffect } from "react";
import { feastEndpoints, useFeastApi } from "../../hooks/useFeastApi";

const FeastResourcesContext = React.createContext({});

export const FeastResourcesContextProvider = ({ project, children }) => {
  const [{ data: entities }, listEntities] = useFeastApi(
    `${feastEndpoints.listEntities}?project=${project}`,
    { method: "GET", muteError: true },
    {},
    false
  );

  const [{ data: featureTables }, listFeatureTables] = useFeastApi(
    `${feastEndpoints.listFeatureTables}?project=${project}`,
    { method: "GET", muteError: true },
    {},
    false
  );

  useEffect(() => {
    console.log("project:", project);
    if (project && project !== "") {
      listEntities();

      listFeatureTables();
    }
  }, [project, listEntities, listFeatureTables]);

  return (
    <FeastResourcesContext.Provider value={{ entities, featureTables }}>
      {children}
    </FeastResourcesContext.Provider>
  );
};

export default FeastResourcesContext;
