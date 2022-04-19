from enum import Enum


class MetricsType(Enum):
    """
    Metrics type to be used for AutoscalingPolicy.

    CONCURRENCY: number of concurrent request handled.
    CPU_UTILIZATION: percentage of CPU utilization.
    MEMORY_UTILIZATION: percentage of Memory utilization.
    RPS: throughput in request per second.
    """
    CONCURRENCY = "concurrency"
    CPU_UTILIZATION = "cpu_utilization"
    MEMORY_UTILIZATION = "memory_utilization"
    RPS = "rps"


class AutoscalingPolicy:
    """
    Autoscaling policy to be used for a deployment.

    """

    def __init__(self, metrics_type: MetricsType, target_value: float):
        self._metrics_type = metrics_type
        self._target_value = target_value

    @property
    def metrics_type(self) -> MetricsType:
        """
        Metrics type to be used for the autoscaling
        :return: MetricsType
        """
        return self._metrics_type

    @property
    def target_value(self) -> float:
        """
        Target metrics value when autoscaling should be performed
        :return: target value
        """
        return self._target_value


RAW_DEPLOYMENT_DEFAULT_AUTOSCALING_POLICY = AutoscalingPolicy(MetricsType.CPU_UTILIZATION, 50)
SERVERLESS_DEFAULT_AUTOSCALING_POLICY = AutoscalingPolicy(MetricsType.CONCURRENCY, 1)
