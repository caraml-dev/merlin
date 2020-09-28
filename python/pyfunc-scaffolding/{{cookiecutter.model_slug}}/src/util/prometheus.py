import re

def safe_prometheus_name(name):
    return re.sub('[^0-9a-zA-Z_]+', '_', name)