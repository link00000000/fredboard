from enum import Enum

import aiohttp

from .Errors import HTTPError, UnauthorizedError, RateLimitError

API_VERSION = 9
BASE_URL = f"https://discord.com/api/v{API_VERSION}"

class _User:
    username: str
    discriminator: str

    def __init__(self, api_response: dict):
        self.username = api_response['username']
        self.discriminator = api_response['discriminator']

class HttpStatusCode(Enum):
    # 2xx
    OK = 200

    # 4xx
    UNAUTHORIZED = 401
    TOO_MANY_REQUESTS = 429

class DiscordClient():
    def __init__(self, token: str):
        self.__token = token

        global_session_headers = {
            "Authorization": token
        }

        self.__session = aiohttp.ClientSession(headers=global_session_headers)

    async def close(self):
        """Cleanup HTTP session."""
        await self.__session.close()

    @staticmethod
    def __raise_http_exception_if_error(response, method: str, route: str):
        """Raise exception if there was an error with HTTP request."""
        if response.status == HttpStatusCode.OK.value:
            return

        if response.status == HttpStatusCode.UNAUTHORIZED.value:
            raise UnauthorizedError()

        if response.status == HttpStatusCode.TOO_MANY_REQUESTS.value:
            raise RateLimitError()

        raise HTTPError({"status": status, "message": f"Unexpected response: {method.upper()} {route} - {response.status}"})

    async def send_message(self, content: str, channel_id: str):
        """Send a message to a Discord channel."""
        route = '/channels/' + channel_id + '/messages'
        create_message_body = {
            "content": content
        }

        async with self.__session.post(BASE_URL + route, json=create_message_body) as response:
            self.__raise_http_exception_if_error(response, 'POST', route)

    async def id(self) -> _User:
        """Get information about currently logged in user"""
        route = '/users/@me'

        async with self.__session.get(BASE_URL + route) as response:
            self.__raise_http_exception_if_error(response, 'GET', route)

            return _User(await response.json())
