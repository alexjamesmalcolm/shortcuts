from typing import TypedDict
from time import sleep
from pydantic import BaseModel
import pydantic
import requests
from tasks_core import define_task


LON_LAT_PATTERN = r"^-?\d+\.\d+,-?\d+\.\d+$"


class TravelTimeInput(BaseModel):
    start_lon_lat: str = pydantic.Field(
        pattern=LON_LAT_PATTERN, examples=["-122.4194,37.7749"]
    )
    end_lon_lat: str = pydantic.Field(
        pattern=LON_LAT_PATTERN, examples=["-118.2437,34.0522"]
    )
    profile: str = "driving"  # OSRM profile: driving, walking, cycling
    max_retries: int = 10
    timeout_seconds: float = 5.0
    retry_pause_seconds: float = 0.2


class TravelTimeResult(TypedDict):
    duration: float
    profile: str


@define_task("travel-time", TravelTimeInput)
def wrapper(inp: TravelTimeInput, progress):
    return travel_time(inp, progress)


def travel_time(inp: TravelTimeInput, progress) -> TravelTimeResult:
    url = (
        f"https://router.project-osrm.org/route/v1/"
        f"{inp.profile}/{inp.start_lon_lat};{inp.end_lon_lat}"
    )
    params = {"overview": "false", "alternatives": "false", "steps": "false"}

    # Try a few times on transient errors
    for attempt in range(1, max(1, inp.max_retries) + 1):
        try:
            response = requests.get(url, params, timeout=inp.timeout_seconds)
            if response.status_code >= 400:
                # Back off on server/client errors and try again
                progress(min(99, int(attempt / max(1, inp.max_retries) * 100)))
                sleep(inp.retry_pause_seconds)
                continue

            data = response.json()
            duration = data["routes"][0]["duration"]
            progress(100)
            return {"duration": float(duration), "profile": inp.profile}
        except requests.Timeout:
            progress(min(99, int(attempt / max(1, inp.max_retries) * 100)))
            sleep(inp.retry_pause_seconds)

    # If all retries exhausted
    raise RuntimeError("Failed to fetch travel time from OSRM after retries")
