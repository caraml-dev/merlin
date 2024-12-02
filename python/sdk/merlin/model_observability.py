from __future__ import annotations

import client
from typing import Optional
from merlin.util import autostr
from dataclasses import dataclass
from dataclasses_json import dataclass_json


@autostr
@dataclass_json
@dataclass
class GroundTruthSource:
    table_urn: str
    event_timestamp_column: str
    source_project: str


@autostr
@dataclass_json
@dataclass
class GroundTruthJob:
    cron_schedule: str
    service_account_secret_name: str
    start_day_offset_from_now: int
    end_day_offset_from_now: int
    cpu_request: Optional[str] = None
    cpu_limit: Optional[str] = None
    memory_request: Optional[str] = None
    memory_limit: Optional[str] = None
    grace_period_day: Optional[int] = None


@autostr
@dataclass_json
@dataclass
class PredictionLogIngestionResourceRequest:
    replica: int
    cpu_request: Optional[str] = None
    memory_request: Optional[str] = None


@autostr
@dataclass_json
@dataclass
class ModelObservability:
    enabled: bool
    ground_truth_source: Optional[GroundTruthSource] = None
    ground_truth_job: Optional[GroundTruthJob] = None
    prediction_log_ingestion_resource_request: Optional[PredictionLogIngestionResourceRequest] = None

    def to_model_observability_request(self) -> client.ModelObservability:
        ground_truth_source = None
        if self.ground_truth_source is not None:
            ground_truth_source = client.GroundTruthSource(
                table_urn=self.ground_truth_source.table_urn,
                event_timestamp_column=self.ground_truth_source.event_timestamp_column,
                source_project=self.ground_truth_source.source_project
            )

        ground_truth_job = None
        if self.ground_truth_job is not None:
            ground_truth_job = client.GroundTruthJob(
                cron_schedule=self.ground_truth_job.cron_schedule,
                service_account_secret_name=self.ground_truth_job.service_account_secret_name,
                start_day_offset_from_now=self.ground_truth_job.start_day_offset_from_now,
                end_day_offset_from_now=self.ground_truth_job.end_day_offset_from_now,
                cpu_request=self.ground_truth_job.cpu_request,
                cpu_limit=self.ground_truth_job.cpu_limit,
                memory_request=self.ground_truth_job.memory_request,
                memory_limit=self.ground_truth_job.memory_limit,
                grace_period_day=self.ground_truth_job.grace_period_day
            )
        
        prediction_log_ingestion_resource_request = None
        if self.prediction_log_ingestion_resource_request is not None:
            prediction_log_ingestion_resource_request = client.PredictionLogIngestionResourceRequest(
                replica=self.prediction_log_ingestion_resource_request.replica,
                cpu_request=self.prediction_log_ingestion_resource_request.cpu_request,
                memory_request=self.prediction_log_ingestion_resource_request.memory_request
            )

        return client.ModelObservability(
            enabled=self.enabled,
            ground_truth_source=ground_truth_source,
            ground_truth_job=ground_truth_job,
            prediction_log_ingestion_resource_request=prediction_log_ingestion_resource_request
        )

    @classmethod
    def from_model_observability_response(cls, response: Optional[client.ModelObservability]) -> Optional[ModelObservability]:
        if response is None:
            return None

        ground_truth_source = None
        if response.ground_truth_source is not None:
            ground_truth_source = cls._ground_truth_source_from_response(response.ground_truth_source)

        ground_truth_job = None
        if response.ground_truth_job is not None:
            ground_truth_job = cls._job_from_response(response.ground_truth_job)

        prediction_log_ingestion_resource_request = None
        if response.prediction_log_ingestion_resource_request is not None:
            prediction_log_ingestion_resource_request = cls._prediction_log_ingestion_resource_request_from_response(response.prediction_log_ingestion_resource_request)

        return ModelObservability(
            enabled=response.enabled,
            ground_truth_source=ground_truth_source,
            ground_truth_job=ground_truth_job,
            prediction_log_ingestion_resource_request=prediction_log_ingestion_resource_request
        )
    
    @classmethod
    def _ground_truth_source_from_response(cls, response: Optional[client.GroundTruthSource]) -> Optional[GroundTruthSource]:
        if response is None:
            return None

        return GroundTruthSource(
            table_urn=response.table_urn,
            event_timestamp_column=response.event_timestamp_column,
            source_project=response.source_project
        )
    
    @classmethod
    def _job_from_response(cls, response: Optional[client.GroundTruthJob]) -> Optional[GroundTruthJob]:
        if response is None:
            return None

        return GroundTruthJob(
            cron_schedule=response.cron_schedule,
            service_account_secret_name=response.service_account_secret_name,
            start_day_offset_from_now=response.start_day_offset_from_now,
            end_day_offset_from_now=response.end_day_offset_from_now,
            cpu_request=response.cpu_request,
            cpu_limit=response.cpu_limit,
            memory_request=response.memory_request,
            memory_limit=response.memory_limit,
            grace_period_day=response.grace_period_day
        )
    
    @classmethod
    def _prediction_log_ingestion_resource_request_from_response(cls, response: Optional[client.PredictionLogIngestionResourceRequest]) -> Optional[PredictionLogIngestionResourceRequest]:
        if response is None:
            return None

        return PredictionLogIngestionResourceRequest(
            replica=response.replica,
            cpu_request=response.cpu_request,
            memory_request=response.memory_request
        )