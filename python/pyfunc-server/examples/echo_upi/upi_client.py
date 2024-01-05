import grpc
import pandas as pd
from caraml.upi.utils import df_to_table
from caraml.upi.v1 import upi_pb2, upi_pb2_grpc


def create_upi_request() -> upi_pb2.PredictValuesRequest:
    target_name = "echo"
    df = pd.DataFrame(
        [[4, 1, "hi"]] * 3,
        columns=["int_value", "int_value_2", "string_value"],
        index=["0000", "1111", "2222"],
    )
    prediction_id = "12345"

    return upi_pb2.PredictValuesRequest(
        target_name=target_name,
        prediction_table=df_to_table(df, "predict"),
        metadata=upi_pb2.RequestMetadata(prediction_id=prediction_id),
    )


if __name__ == "__main__":
    channel = grpc.insecure_channel(f"localhost:8080")
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)

    request = create_upi_request()
    response = stub.PredictValues(request=request)
    print(response)
