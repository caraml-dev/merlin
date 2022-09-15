import time
import requests


def wait_server_ready(url, timeout_second=10, tick_second=2):
    ellapsed_second = 0
    while (ellapsed_second < timeout_second):
        try:
            resp = requests.get(url)
            if resp.status_code == 200:
                return
        except Exception as e:
            print(f"{url} is not ready: {e}")

        time.sleep(tick_second)
        ellapsed_second += tick_second

    if ellapsed_second >= timeout_second:
        raise TimeoutError("server is not ready within specified timeout duration")
