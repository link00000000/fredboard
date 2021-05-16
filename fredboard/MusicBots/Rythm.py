from .AbstractMusicBot import AbstractMusicBot, AbstractMusicBotConfig
from ..Discord import DiscordClient

_BOT_ID = "rythm"

class RythmMusicBotConfig(AbstractMusicBotConfig):
    id = _BOT_ID
    name: str = _BOT_ID
    command_prefix: str = "!"

class RythmMusicBot(AbstractMusicBot):
    id = _BOT_ID

    def __init__(self, discord_client: DiscordClient, config: RythmMusicBotConfig):
        super().__init__(discord_client, config)

    def command(self, command_name, *args):
        command_str = self.config.command_prefix + command_name

        if args is not None and len(args) > 0:
            command_str += " " + " ".join(args)

        return command_str

    async def start_audio(self, audio_url: str):
        await self.send_message_if_connected(self.command("play", audio_url))

    async def stop_audio(self):
        await self.send_message_if_connected(self.command("disconnect"))
