from abc import abstractmethod
from typing import Union

import grpc
import numpy
import pandas
from caraml.upi.v1 import upi_pb2
from mlflow.pyfunc import PythonModel

from merlin.protocol import Protocol

PYFUNC_EXTRA_ARGS_KEY = "__EXTRA_ARGS__"
PYFUNC_MODEL_INPUT_KEY = "__INPUT__"
PYFUNC_PROTOCOL_KEY = "__PROTOCOL__"
PYFUNC_GRPC_CONTEXT = "__GRPC_CONTEXT__"


class PyFuncModel(PythonModel):
    def load_context(self, context):
        """
        Override method of PythonModel `load_context`. This method load artifacts from context
        that can be used in predict function. This method is called by internal MLflow when an MLflow
        is loaded.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        """
        self.initialize(context.artifacts)
        self._use_kwargs_infer = True

    def predict(self, context, model_input):
        """
        Implementation of PythonModel `predict` method. This method evaluates model input and produces model output.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        :param model_input: A pyfunc-compatible input for the model to evaluate.
        """
        extra_args = model_input.get(PYFUNC_EXTRA_ARGS_KEY, {})
        input = model_input.get(PYFUNC_MODEL_INPUT_KEY, {})
        protocol = model_input.get(PYFUNC_PROTOCOL_KEY, Protocol.HTTP_JSON)

        if protocol == Protocol.HTTP_JSON:
            if extra_args is not None:
                return self._do_http_predict(input, **extra_args)
            return self._do_http_predict(input)
        elif protocol == Protocol.UPI_V1:
            grpc_context = model_input.get(PYFUNC_GRPC_CONTEXT, None)
            return self.upiv1_infer(input, grpc_context)
        else:
            raise NotImplementedError(f"protocol {protocol} is not supported")

    def _do_http_predict(self, model_input, **kwargs):
        if self._use_kwargs_infer:
            try:
                return self.infer(model_input, **kwargs)
            except TypeError as e:
                if "infer() got an unexpected keyword argument" in str(e):
                    print(
                        'Fallback to the old infer() method, got TypeError exception: {}'.format(e))
                    self._use_kwargs_infer = False
                else:
                    raise e

        return self.infer(model_input)

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.

        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    @abstractmethod
    def infer(self, request: dict, **kwargs) -> dict:
        """
        Do inference.

        This method MUST be implemented by concrete implementation of
        PyFuncModel to support HTTP_JSON protocol.
        This method accept 'request' which is the body content of incoming
        request.
        Implementation should return inference a json object of response.

        :param request: Dictionary containing incoming request body content
        :param **kwargs: See below.

        :return: Dictionary containing response body

        :keyword arguments:
        * headers (dict): Dictionary containing incoming HTTP request headers
        """
        raise NotImplementedError("infer is not implemented")

    @abstractmethod
    def upiv1_infer(self, request: upi_pb2.PredictValuesRequest,
                    context: grpc.ServicerContext) -> upi_pb2.PredictValuesResponse:
        """
        Do inference.

        This method MUST be implemented by concrete implementation of PyFunModel to support UPI_V1 protocol.
        The method accept request in for of PredictValuesRequest proto and should return PredictValuesResponse response proto.

        :param request: Inference request as PredictValuesRequest
        :param context: grpc context
        :return: Prediction result as PredictValuesResponse proto
        """
        raise NotImplementedError("upiv1_infer is not implemented")



class PyFuncV2Model(PythonModel):
    def load_context(self, context):
        """
        Override method of PythonModel `load_context`. This method load artifacts from context
        that can be used in predict function. This method is called by internal MLflow when an MLflow
        is loaded.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        """
        self.initialize(context.artifacts)

    def predict(self, context, model_input):
        """
        Implementation of PythonModel `predict` method. This method evaluates model input and produces model output.

        :param context: A :class:`~PythonModelContext` instance containing artifacts that the model
                        can use to perform inference
        :param model_input: A pyfunc-compatible input for the model to evaluate.
        """
        return self.infer(model_input)

    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.

        :param artifacts: dictionary of artifacts passed to log_model method
        """
        pass

    def infer(self, model_input: pandas.DataFrame) -> Union[numpy.ndarray,
                                                            pandas.Series,
                                                            pandas.DataFrame]:
        """
        Infer method is the main method that will be called when calculating
        the inference result for both online prediction and batch
        prediction. The method accepts pandas Dataframe and returns either
        another panda Dataframe / pandas Series / ndarray of the same length
        as the input. In the batch prediction case the model_input will
        contain an arbitrary partition of the whole dataset that the user
        defines as the data source. As such, it is advisable not to do
        aggregation within the infer method, as it will be incorrect since
        it will only apply to the partition in contrary to the whole dataset.

        :param model_input: input to the model (pandas.DataFrame)
        :return: inference result as numpy.ndarray or pandas.Series or pandas.DataFrame

        """
        raise NotImplementedError("infer is not implemented")

    def preprocess(self, request: dict) -> pandas.DataFrame:
        """
        Preprocess incoming request into a pandas Dataframe that will be
        passed to the infer method.
        This method will not be called during batch prediction.

        :param request: dictionary representing the incoming request body
        :return: pandas.DataFrame that will be passed to infer method
        """
        raise NotImplementedError("preprocess is not implemented")

    def postprocess(self, model_result: Union[numpy.ndarray,
                                              pandas.Series,
                                              pandas.DataFrame]) -> dict:
        """
        Postprocess prediction result returned by infer method into
        dictionary representing the response body of the model.
        This method will not be called during batch prediction.

        :param model_result: output of the model's infer method
        :return: dictionary containing the response body
        """
        raise NotImplementedError("postprocess is not implemented")

    def raw_infer(self, request: dict) -> dict:
        """
        Do inference
        This method MUST be implemented by concrete implementation of
        PyFuncV2Model.
        This method accept 'request' which is the body content of incoming
        request.
        This method will not be called during batch prediction.

        Implementation should return inference a json object of response.

        :param request: Dictionary containing incoming request body content
        :return: Dictionary containing response body
        """
        raise NotImplementedError("raw_infer is not implemented")
