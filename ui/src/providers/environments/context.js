import React from "react";
import mocks from "../../mocks";

const EnvironmentsContext = React.createContext([]);

export const EnvironmentsContextProvider = ({ children }) => {
  // const [{ data: environments }] = useMerlinApi("/environments", {}, []);

  const environments = mocks.environmentList;

  return (
    <EnvironmentsContext.Provider value={environments}>
      {children}
    </EnvironmentsContext.Provider>
  );
};

export default EnvironmentsContext;
