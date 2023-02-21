import logging
import socket
import time
from threading import Thread

from prometheus_client import push_to_gateway

from pyfuncserver.config import Config


def start_metrics_pusher(push_gateway, registry, target_info, interval_sec):
    """
    Start periodic job to push metrics to prometheus push gatewat
    """
    logging.info(f"starting metrics pusher, url: {push_gateway} with interval {interval_sec} s")
    daemon = Thread(target=push_metrics, args=(push_gateway, registry, interval_sec, target_info),
                    daemon=True, name='metrics_push')
    daemon.start()


def push_metrics(gateway_url, registry, interval_sec, grouping_keys):
    """
    push metrics to prometheus push gateway every interval_sec
    Should be called in separate thread
    """
    while True:
        push_to_gateway(gateway_url, "merlin_pyfunc_upi", registry=registry, grouping_key=grouping_keys)
        time.sleep(interval_sec)


def labels(config: Config):
    """
    Labels to be added to all metrics
    """
    return {
        "merlin_model_name": config.model_manifest.model_name,
        "merlin_model_version": config.model_manifest.model_version,
        "host": socket.getfqdn()
    }
