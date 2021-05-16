from enum import Enum
import asyncio

import aiohttp
import discord as discordpy

from .Errors import HTTPError, UnauthorizedError, RateLimitError

API_VERSION = 9
BASE_URL = f"https://discord.com/api/v{API_VERSION}"

class _User:
    username: str
    discriminator: str

    def __init__(self, api_response: dict):
        self.username = api_response['username']
        self.discriminator = api_response['discriminator']

class _TextChannel:
    name: str
    guild_id: str

    def __init__(self, api_response: dict):
        self.name = api_response['name']
        self.guild_id = api_response['guild_id']

class _Guild:
    name: str
    id: str

    def __init__(self, api_response: dict):
        self.name = api_response['name']
        self.id = api_response['id']

class HttpStatusCode(Enum):
    # 2xx
    OK = 200

    # 4xx
    UNAUTHORIZED = 401
    TOO_MANY_REQUESTS = 429

class DiscordClient():
    class _DiscordPyClient(discordpy.Client):
        def run(self, token: str):
            self.task = asyncio.create_task(self.start(token, bot=False))

        async def stop(self):
            if not self.is_closed():
                await self.close()

    def __init__(self, token: str):
        self.__token = token

        global_session_headers = {
            "Authorization": token
        }

        self.__session = aiohttp.ClientSession(headers=global_session_headers)
        self.__discordpy_client = self._DiscordPyClient()
        self.__discordpy_client.run(token)

    async def close(self):
        """Cleanup HTTP session."""
        await self.__session.close()
        await self.__discordpy_client.stop()

    @staticmethod
    def __raise_http_exception_if_error(response, method: str, route: str):
        """Raise exception if there was an error with HTTP request."""
        if response.status == HttpStatusCode.OK.value:
            return

        if response.status == HttpStatusCode.UNAUTHORIZED.value:
            raise UnauthorizedError()

        if response.status == HttpStatusCode.TOO_MANY_REQUESTS.value:
            raise RateLimitError()

        raise HTTPError({"status": response.status, "message": f"Unexpected response: {method.upper()} {route} - {response.status}"})

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

    async def text_channel(self, channel_id: str) -> _TextChannel:
        """Get information about text channel with channel ID"""
        route = '/channels/' + channel_id

        async with self.__session.get(BASE_URL + route) as response:
            self.__raise_http_exception_if_error(response, 'GET', route)

            return _TextChannel(await response.json())

    async def guild(self, guild_id: str) -> _Guild:
        """Get information about guild with guild ID"""
        route = '/guilds/' + guild_id

        async with self.__session.get(BASE_URL + route) as response:
            self.__raise_http_exception_if_error(response, 'GET', route)

            return _Guild(await response.json())

    async def connected_voice_guild_id(self) -> _Guild:
        """Get information about current voice channel guild"""
        for guild in self.__discordpy_client.guilds:
            if guild.me.voice is not None:
                return _Guild({
                    'id': str(guild.id),
                    'name': str(guild.name)
                })

