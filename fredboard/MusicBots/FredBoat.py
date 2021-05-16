from .AbstractMusicBot import AbstractMusicBot, AbstractMusicBotConfig
from ..Discord import DiscordClient
from ..Logger import logger
from ..Errors import RateLimitError, UnauthorizedError, HTTPError 

_BOT_ID = "fredboat"

class FredboatMusicBotConfig(AbstractMusicBotConfig):
    id = _BOT_ID
    name: str = _BOT_ID
    command_prefix: str = ";;"

class FredboatMusicBot(AbstractMusicBot):
    id = _BOT_ID

    def __init__(self, discord_client: DiscordClient, config: FredboatMusicBotConfig):
        super().__init__(discord_client, config)

    async def start_audio(self, audio_url: str):
        await self.send_message_if_connected(self.config.command_prefix + "play " + audio_url)

    async def stop_audio(self):
        await self.send_message_if_connected(self.config.command_prefix + "stop")

