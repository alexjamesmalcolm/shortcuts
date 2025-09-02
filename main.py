from typing import Annotated

from fastapi import FastAPI, Query
from functions import get_travel_time

app = FastAPI()


@app.get("/travel-time")
def handle_get_travel_time(
    start: Annotated[
        str,
        Query(
            description='The location of the starting position. In the format of "lon,lat" as '
            "floats without a space after the comma.",
            pattern="^-?\d+\.\d+,-?\d+\.\d+$",
        ),
    ],
    end: Annotated[
        str,
        Query(
            description='The location of the ending position. In the format of "lon,lat" as floats '
            "without a space after the comma.",
            pattern="^-?\d+\.\d+,-?\d+\.\d+$",
        ),
    ],
):
    return get_travel_time(start, end)
