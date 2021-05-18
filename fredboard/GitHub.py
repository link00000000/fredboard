import aiohttp
from pydantic import BaseModel, Extra

class Asset(BaseModel, extra=Extra.ignore):
    url: str
    name: str
    browser_download_url: str

class Release(BaseModel, extra=Extra.ignore):
    url: str
    tag_name: str
    assets: list[Asset]

class GitHub():
    _BASE_URL = "https://api.github.com"

    def __init__(self, repository: str):
        self.__session = aiohttp.ClientSession()
        self.__repository = repository

    async def __aenter__(self):
        return self

    async def __aexit__(self, *args):
        await self.close()

    async def close(self):
        await self.__session.close()

    async def latest_release(self) -> Release:
        response = await self.__session.get(self._BASE_URL + '/repos/' + self.__repository + '/releases/latest')
        return Release(**await response.json())
