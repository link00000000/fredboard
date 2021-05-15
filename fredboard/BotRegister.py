import asyncio
from typing import Tuple

from .Discord import DiscordClient
from .MusicBots.Types import get_music_bot_type_by_name
from .MusicBots.AbstractMusicBot import AbstractMusicBotConfig, AbstractMusicBot

class BotRegister:
    @staticmethod
    async def initialize_music_bots_from_config(music_bots_configs: list[AbstractMusicBotConfig], discord: DiscordClient) -> "Tuple(list[AbstractMusicBot], list[Exception])":
        results = await asyncio.gather(*[BotRegister.initialize_music_bot_from_config(bot_config, discord) for bot_config in music_bots_configs], return_exceptions=True)
        
        exceptions = []
        music_bots = []
        for result in results:
            if isinstance(result, Exception):
                exceptions.append(result)
            else:
                music_bots.append(result)
            
        return music_bots, exceptions

    @staticmethod
    async def initialize_music_bot_from_config(music_bot_config: AbstractMusicBotConfig, discord: DiscordClient) -> AbstractMusicBot:
        BotType, BotConfigType = get_music_bot_type_by_name(music_bot_config.name)
        return BotType(discord, BotConfigType(**music_bot_config.dict()))

