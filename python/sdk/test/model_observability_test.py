import pytest
import client
from merlin.model_observability import ModelObservability, GroundTruthSource, GroundTruthJob, PredictionLogIngestionResourceRequest

@pytest.mark.unit
@pytest.mark.parametrize(
    "response,expected", [
        (
            client.ModelObservability(
                enabled=True,
                ground_truth_source=client.GroundTruthSource(
                    table_urn="table_urn",
                    event_timestamp_column="event_timestamp_column",
                    source_project="dwh_project"
                ),
                ground_truth_job=client.GroundTruthJob(
                    cron_schedule="cron_schedule",
                        service_account_secret_name="service_account_secret_name",
                        start_day_offset_from_now=1,
                        end_day_offset_from_now=1,
                        cpu_request="cpu_request",
                        cpu_limit="cpu_limit",
                        memory_request="memory_request",
                        memory_limit="memory_limit",
                        grace_period_day=1
                ),
                prediction_log_ingestion_resource_request=client.PredictionLogIngestionResourceRequest(
                    replica=1,
                    cpu_request="1",
                    memory_request="1Gi"
                )
            ),
            ModelObservability(
                enabled=True,
                ground_truth_source=GroundTruthSource(
                    table_urn="table_urn",
                    event_timestamp_column="event_timestamp_column",
                    source_project="dwh_project"
                ),
                ground_truth_job=GroundTruthJob(
                    cron_schedule="cron_schedule",
                    service_account_secret_name="service_account_secret_name",
                    start_day_offset_from_now=1,
                    end_day_offset_from_now=1,
                    cpu_request="cpu_request",
                    cpu_limit="cpu_limit",

                    memory_request="memory_request",
                    memory_limit="memory_limit",
                    grace_period_day=1
                ),
                prediction_log_ingestion_resource_request=PredictionLogIngestionResourceRequest(
                    replica=1,
                    cpu_request="1",
                    memory_request="1Gi"
                )
            )
        )
    ]
)
def test_from_model_observability_response(response, expected):
    assert ModelObservability.from_model_observability_response(response) == expected
    assert expected.to_model_observability_request() == response