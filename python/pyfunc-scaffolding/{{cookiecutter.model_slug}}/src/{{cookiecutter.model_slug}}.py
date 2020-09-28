from merlin.model import PyFuncModel
import src.constant as constant

{%- from "deployment.py" import model_class with context %}

class {{ model_class() }}(PyFuncModel):

    def __init__(self, namespace, model_name):
        self._namespace = namespace
        self._model_name = model_name

    def initialize(self, artifacts: dict):
        """
        Implementation of PyFuncModel can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_model method
        """
        self._artifacts = artifacts
        self._model = self.__load_model(artifacts[constant.ARTIFACT_MODEL_PATH])

    def __load_model(self, model_path):
        """
        Load model from artifacts
        """
        return None

    def infer(self, request: dict) -> dict:
        """
        Do inference
        This method MUST be implemented by concrete implementation of
        PyFuncModel.
        This method accept 'request' which is the body content of incoming
        request.
        Implementation should return inference a json object of response.
        :param request: Dictionary containing incoming request body content
        :return: Dictionary containing response body
        """
        if not self.__is_valid_request(request):
            return self.__invalid_request_handler(error_message="invalid request")

        # Data transformation and prediction here

        return self.__success_response_body()

    def __is_valid_request(self, request: dict) -> bool:
        """
        Determine whether request if valid or not
        :param request:
        :return:
        """
        return True

    def __success_response_body(self, **kwargs) -> dict:
        """
        Return response body for success response
        :param kwargs:
        :return:
        """
        return self.__build_response_body(**kwargs)

    def __invalid_request_handler(self, **kwargs) -> dict:
        """
        Return response body for non success response
        :return:
        """
        return self.__build_response_body(**kwargs)

    def __build_response_body(self, **kwargs) -> dict:
        """
        Construct response body based on kwargs parameter
        :param kwargs: key value
        :return:
        """
        response = {}
        for key, value in kwargs.items():
            response[key] = value

        return response
