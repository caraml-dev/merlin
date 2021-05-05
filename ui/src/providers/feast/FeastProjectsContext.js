import React from "react";
import { feastEndpoints, useFeastApi } from "../../hooks/useFeastApi";

const FeastProjectsContext = React.createContext([]);

export const FeastProjectsContextProvider = ({ children }) => {
  const [{ data: projects }] = useFeastApi(
    feastEndpoints.listProjects,
    { method: "POST", muteError: true },
    {},
    true
  );

  return (
    <FeastProjectsContext.Provider value={projects}>
      {children}
    </FeastProjectsContext.Provider>
  );
};

export default FeastProjectsContext;
