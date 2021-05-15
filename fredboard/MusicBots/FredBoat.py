from .AbstractMusicBot import AbstractMusicBot, AbstractMusicBotConfig
from ..Discord import DiscordClient
from ..Logger import logger
from ..Errors import RateLimitError, UnauthorizedError, HTTPError 

class FredboatMusicBotConfig(AbstractMusicBotConfig):
    id = "fredboat"
    name: str = "fredboat"
    command_prefix: str = ";;"

class FredboatMusicBot(AbstractMusicBot):
    id = "fredboat"

    def __init__(self, discord_client: DiscordClient, config: FredboatMusicBotConfig):
        self.config = config
        self.discord_client = discord_client

    async def __send_message(self, message):
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
        
    async def start_audio(self, audio_url: str):
        await self.__send_message(self.config.command_prefix + "play " + audio_url)

    async def stop_audio(self):
        await self.__send_message(self.config.command_prefix + "stop")

