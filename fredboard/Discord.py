from enum import Enum

import aiohttp

API_VERSION = 9
BASE_URL = f"https://discord.com/api/v{API_VERSION}"

class HttpStatusCode(Enum):
    # 2xx
    OK = 200

    # 4xx
    UNAUTHORIZED = 401
    TOO_MANY_REQUESTS = 429

class HTTPError(RuntimeError):
    pass

class UnauthorizedError(HTTPError):
    pass

class RateLimitError(HTTPError):
    pass

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
            raise Unauthorized()

        if response.status == HttpStatusCode.TOO_MANY_REQUESTS.value:
            raise RateLimitError()

        raise HTTPError(f"Unexpected response: {method.upper()} {route} - {response.status}")

    async def send_message(self, content: str, channel_id: str):
        """Send a message to a Discord channel."""
        route = '/channels/' + channel_id + '/messages'
        create_message_body = {
            "content": content
        }

        async with self.__session.post(BASE_URL + route, json=create_message_body) as response:
            self.__raise_http_exception_if_error(response, 'POST', route)

