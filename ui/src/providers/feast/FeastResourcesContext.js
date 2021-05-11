import React, { useEffect } from "react";
import { feastEndpoints, useFeastApi } from "../../hooks/useFeastApi";

const FeastResourcesContext = React.createContext({});

export const FeastResourcesContextProvider = ({ project, children }) => {
  const [{ data: entities }, listEntities] = useFeastApi(
    feastEndpoints.listEntities,
    { method: "POST", muteError: true },
    {},
    false
  );

  const [{ data: featureTables }, listFeatureTables] = useFeastApi(
    feastEndpoints.listFeatureTables,
    { method: "POST", muteError: true },
    {},
    false
  );

  useEffect(() => {
    if (project && project !== "") {
      listEntities({
        body: JSON.stringify({ filter: { project: project } })
      });

      listFeatureTables({
        body: JSON.stringify({ filter: { project: project } })
      });
    }
  }, [project, listEntities, listFeatureTables]);

  return (
    <FeastResourcesContext.Provider value={{ entities, featureTables }}>
      {children}
    </FeastResourcesContext.Provider>
  );
};

export default FeastResourcesContext;
