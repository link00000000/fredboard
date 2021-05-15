from typing import Type, Tuple

from .AbstractMusicBot import AbstractMusicBot, AbstractMusicBotConfig

# Lookup Bot/Config pairs using the string name
__music_bot_types = dict[str, (Type[AbstractMusicBot], Type[AbstractMusicBotConfig])]()

def get_music_bot_type_by_name(name: str) -> "Tuple(Type[AbstractMusicBot], Type[AbstractMusicBotConfig])":
    # Populate cache on first run
    if len(__music_bot_types) == 0:
        configs = AbstractMusicBotConfig.__subclasses__()
        for bot in AbstractMusicBot.__subclasses__():
            for config in configs:
                if bot.id == config.id:
                    __music_bot_types[bot.id] = (bot, config)
                    break
            
            if bot.id not in __music_bot_types:
                raise TypeError(f"Bot type {bot.__name__} is not a valid type.")
        
    if name in __music_bot_types:
        return __music_bot_types[name]
    
    raise KeyError("No music bot with name " + name)
    
