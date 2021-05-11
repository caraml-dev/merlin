import React from "react";
import { appConfig } from "../../config";

const DockerRegistriesContext = React.createContext([]);

export const DockerRegistriesContextProvider = ({ children }) => {
  const registries = [
    ...appConfig.dockerRegistries.map(registry => ({
      value: registry,
      inputDisplay: registry
    })),
    {
      value: "docker.io",
      inputDisplay: "Docker Hub"
    }
  ];

  return (
    <DockerRegistriesContext.Provider value={registries}>
      {children}
    </DockerRegistriesContext.Provider>
  );
};

export default DockerRegistriesContext;
