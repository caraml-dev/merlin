import React, { useContext } from "react";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler
} from "@caraml-dev/ui-lib";
import {PredictionLoggerPanel} from "../components/transformer/components/PredictionLoggerPanel"


export const PredictionLoggerStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  const predictionLogger = get(data, "logger.prediction")
  return (
    <PredictionLoggerPanel 
        predictionLogger={predictionLogger}
        onChangeHandler={onChange("logger.prediction")}
        errors={get(errors, "logger.prediction")}
    />
  );
};
