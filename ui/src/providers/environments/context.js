import React from "react";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const EnvironmentsContext = React.createContext([]);

export const EnvironmentsContextProvider = ({ children }) => {
  const [{ data: environments }] = useMerlinApi("/environments", {}, []);

  console.log("___environments", environments);

  return (
    <EnvironmentsContext.Provider value={environments}>
      {children}
    </EnvironmentsContext.Provider>
  );
};

export default EnvironmentsContext;
