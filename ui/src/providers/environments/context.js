import React from "react";
import { useMerlinApi } from "../../hooks/useMerlinApi";
// import evn from mock
import mocks from "../../mocks";

const EnvironmentsContext = React.createContext([]);

export const EnvironmentsContextProvider = ({ children }) => {
  // const [{ data: environments }] = useMerlinApi("/environments", {}, []);

  const environments = mocks.environmentList;

  console.log("___environments", environments);

  return (
    <EnvironmentsContext.Provider value={environments}>
      {children}
    </EnvironmentsContext.Provider>
  );
};

export default EnvironmentsContext;
