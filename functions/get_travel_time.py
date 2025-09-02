from time import sleep
import requests


def get_travel_time(start_lon_lat: str, end_lon_lat: str) -> float:
    profile = "driving"
    url = f"https://router.project-osrm.org/route/v1/{profile}/{start_lon_lat};{end_lon_lat}"
    params = {"overview": "false", "alternatives": "false", "steps": "false"}
    pause_duration = 0.1
    while True:
        try:
            response = requests.get(url, params, timeout=5)
            if response.status_code >= 400:
                sleep(pause_duration)
            else:
                return response.json()["routes"][0]["duration"]
        except requests.Timeout:
            sleep(pause_duration)
