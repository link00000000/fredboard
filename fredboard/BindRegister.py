from typing import Callable
import asyncio

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys
from .Settings import KeyBind, StopKeyBind, QuitKeyBind, AudioKeyBind
from .MusicBots.AbstractMusicBot import AbstractMusicBot
from .MusicBots.Types import get_music_bot_type_by_name

class BindRegiser():
    def __init__(self, keybinds: list[KeyBind], music_bots: list[AbstractMusicBot]):
        self.on_quit_callbacks = set[Callable]()

        bindings = []
        for keybind in keybinds:
            if isinstance(keybind, AudioKeyBind):
                bindings.append([keybind.sequence, None, self.__create_bind(music_bots, keybind.audio)])
            
            elif isinstance(keybind, StopKeyBind):
                bindings.append([keybind.sequence, None, self.__create_stop_bind(music_bots)])

            elif isinstance(keybind, QuitKeyBind):
                bindings.append([keybind.sequence, None, self.__create_quit_bind()])

            else:
                raise TypeError("Unable to register hotkey of type " + keybind.__class__.__name__)

        hotkeys.clear_hotkeys()
        hotkeys.register_hotkeys(bindings)

    def __enter__(self):
        hotkeys.start_checking_hotkeys()
        return self

    def __exit__(self, *args):
        hotkeys.stop_checking_hotkeys()

    def __create_bind(self, music_bots: list[AbstractMusicBot], url: str):
        async def play_audio():
            await asyncio.gather(*[music_bot.start_audio(url) for music_bot in music_bots])

        return play_audio

    def __create_stop_bind(self, music_bots: list[AbstractMusicBot]):
        async def stop_audio():
            await asyncio.gather(*[music_bot.stop_audio() for music_bot in music_bots])

        return stop_audio

    def __create_quit_bind(self):
        def quit_bind():
            for callback in self.on_quit_callbacks:
                callback()

        return quit_bind

    def on_quit(self, func: Callable):
        """Register callback function for when quit hotkey is pressed."""
        self.on_quit_callbacks.add(func)
        return func
