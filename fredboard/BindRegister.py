from typing import Callable
import asyncio

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys
from .Settings import KeyBind, StopKeyBind, QuitKeyBind, AudioKeyBind
from .MusicBots.AbstractMusicBot import AbstractMusicBot
from .MusicBots.Types import get_music_bot_type_by_name
from .Logger import logger
from .Youtube import YoutubeAPI

class BindRegiser():
    def __init__(self, keybinds: list[KeyBind], music_bots: list[AbstractMusicBot]):
        self.on_quit_callbacks = set[Callable]()

        self.stop_binding: StopKeyBind
        self.quit_binding: QuitKeyBind
        self.audio_bindings = list[AudioKeyBind]()

        bindings = []
        for keybind in keybinds:
            if isinstance(keybind, AudioKeyBind):
                self.audio_bindings.append(keybind)
                bindings.append([keybind.sequence, None, self.__create_bind(music_bots, keybind.audio)])
            
            elif isinstance(keybind, StopKeyBind):
                self.stop_binding = keybind
                bindings.append([keybind.sequence, None, self.__create_stop_bind(music_bots)])

            elif isinstance(keybind, QuitKeyBind):
                self.quit_binding = keybind
                bindings.append([keybind.sequence, None, self.__create_quit_bind()])

            else:
                raise TypeError("Unable to register hotkey of type " + keybind.__class__.__name__)

        hotkeys.clear_hotkeys()
        hotkeys.register_hotkeys(bindings)

    async def __aenter__(self):
        hotkeys.start_checking_hotkeys()

        # Log registered hotkeys
        async with YoutubeAPI() as youtube:
            logger.info("Registered global keybinds:")
            logger.info("\t" + "+".join(self.stop_binding.sequence) + " - Stop")
            logger.info("\t" + "+".join(self.quit_binding.sequence) + " - Quit")
            for bind in self.audio_bindings:
                if youtube.is_youtube_video(bind.audio):
                    logger.info("\t" + "+".join(bind.sequence) + ' - YouTube: ' + await youtube.video_title(bind.audio))
                else:
                    logger.info("\t" + "+".join(bind.sequence) + ' - ' + bind.audio)

        return self

    async def __aexit__(self, *args):
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
