from typing import Dict, List, Optional
from itertools import permutations
import math
import pydantic
from tasks_core import define_task

from .travel_time import travel_time, TravelTimeInput


class Location(pydantic.BaseModel):
    lat: float
    lon: float
    address: str


class OptimalRouteInput(pydantic.BaseModel):
    origin: Location
    destination: Location
    stops: List[Location]


@define_task("optimal-route", OptimalRouteInput)
def optimal_route(inp: OptimalRouteInput, progress) -> dict:
    locations: List[Location] = [inp.origin, inp.destination] + inp.stops

    travel_times: Dict[str, Dict[str, float]] = {}
    location_permutations = list(permutations(locations, 2))
    routes = list(permutations(locations))
    total_steps = len(location_permutations) + len(routes)
    step_count: int = 0

    def mark_step_done():
        nonlocal step_count
        step_count = step_count + 1
        progress(step_count / total_steps)

    for start, end in location_permutations:
        if (
            start == inp.destination
            or end == inp.origin
            or (start == inp.origin and end == inp.destination)
        ):
            mark_step_done()
            continue
        if start.address not in travel_times:
            travel_times[start.address] = {}
        travel_times[start.address][end.address] = travel_time(
            TravelTimeInput(
                start_lon_lat=f"{start.lon},{start.lat}",
                end_lon_lat=f"{end.lon},{end.lat}",
            ),
            lambda percent: None,
        )["duration"]
        mark_step_done()

    best_travel_time: float = math.inf
    best_route: Optional[List[Location]] = None
    for route in routes:
        if route[0] != inp.origin or route[-1] != inp.destination:
            mark_step_done()
            continue
        route_travel_time: float = 0.0
        for i in range(len(route) - 1):
            start = route[i]
            end = route[i + 1]
            route_travel_time += travel_times[start.address][end.address]
        if route_travel_time < best_travel_time:
            best_travel_time = route_travel_time
            best_route = list(route)
        mark_step_done()
    if not best_route:
        raise Exception("Unable to find the best route")

    return {
        "travel_times": travel_times,
        "best_route": [location.model_dump_json() for location in best_route],
    }
