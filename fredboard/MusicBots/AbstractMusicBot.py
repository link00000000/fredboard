from abc import ABC, ABCMeta, abstractmethod
from typing import ClassVar

import pydantic
from pydantic import BaseModel

from ..Discord import DiscordClient
from ..Logger import logger

class AbstractMusicBotConfig(BaseModel, ABC):
    id: ClassVar[str]
    name: str
    channel_id: str

    class Config:
        extra = pydantic.Extra.allow

class AbstractMusicBot(metaclass=ABCMeta):
    """
    @NOTE Make sure to import any subclasses into __init__.py so that
          it will be able to be reflexivly referenced.
    """
    id: ClassVar[str]

    def __init__(self, discord_client: DiscordClient, config: AbstractMusicBotConfig):
        self.discord_client = discord_client
        self.config = config

    async def send_message(self, message: str):
        logger.debug(f"Sending message '{message}' to channel with id '{self.config.channel_id}'")
        try:
            await self.discord_client.send_message(message, self.config.channel_id)

        except RateLimitError:
            logger.error("Too many requests made too quickly. Try again later.")

        except UnauthorizedError:
            logger.error("Invalid login token. Did you set your login token in config.json?")
            exit()

        except HTTPError as error:
            if error.status == 400:
                logger.error("Bad request. Did you set your channel id in config.json?")

            else:
                raise error

    async def send_message_if_connected(self, message: str):
        connected_guild = await self.discord_client.connected_voice_guild_id()
        text_channel = await self.discord_client.text_channel(self.config.channel_id)
        if connected_guild is not None and connected_guild.id == text_channel.guild_id:
            await self.send_message(message)

    @abstractmethod
    async def start_audio(self, url: str):
        raise NotImplemented()
    
    @abstractmethod
    async def stop_audio(self):
        raise NotImplemented()

