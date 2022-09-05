import enum


class Protocol(enum.Enum):
    """
    Model deployment protocol.

    HTTP_JSON = Deploy model and expose HTTP server
    UPI_V1 = Deploy model and expose gRPC server compatible with universal-prediction-interface (https://github.com/caraml-dev/universal-prediction-interface)
    """
    HTTP_JSON = "HTTP_JSON"
    UPI_V1 = "UPI_V1"
