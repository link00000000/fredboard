from .AbstractMusicBot import AbstractMusicBot, AbstractMusicBotConifg
from ..Discord import DiscordClient
from ..Logger import logger

class FredboatMusicBotConfig(AbstractMusicBotConifg):
    name = "fredboat"
    command_prefix: str = ";;"

class FredboatMusicBot(AbstractMusicBot):
    def __init__(self, discord_client: DiscordClient, channel_id: str, command_prefix = ";;"):

        self.discord_client = discord_client
        self.channel_id = channel_id
        self.command_prefix = command_prefix

    async def __send_message(self, message):
        try:
            await self.discord_client.send_message(message, self.channel_id)

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
        await self.__send_message(self.command_prefix + "play " + audio_url)

    async def stop_audio(self):
        await self.__send_message(self.command_prefix + "stop")

