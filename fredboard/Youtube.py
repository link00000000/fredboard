from urllib.parse import urlencode
import re

import aiohttp

from .Errors import HTTPError

_youtube_video_matcher = re.compile(r"^https?://(www\.)?youtube\.com/watch\?v=")

BASE_VIDEO_URL = "https://www.youtube.com/watch?v="
BASE_URL = "https://www.youtube.com/oembed"

class YoutubeAPI:
    def __init__(self):
        self.__session = aiohttp.ClientSession()
        self.__cache = dict[str, dict]()

    async def __aenter__(self):
        return self

    async def __aexit__(self, *args):
        await self.__session.close()

    async def __fetch_video_data(self, video_url: str) -> dict:
        if video_url in self.__cache:
            return self.__cache[video_url]

        params = urlencode({
            "format": "json",
            "url": video_url
        })

        async with self.__session.get(BASE_URL + "?" + params) as response:
            if response.status != 200:
                raise HTTPError(f"Unexpected response: GET {BASE_URL + '?' + params} - {response.status}")

            data = await response.json()
            self.__cache[video_url] = data

            return data

    async def video_title(self, video_url: str) -> str:
        return (await self.__fetch_video_data(video_url))["title"]

    def is_youtube_video(self, video_url: str) -> bool:
        return re.match(r"^https?://(www\.)?youtube\.com/watch\?v=", video_url) is not None
