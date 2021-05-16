from __future__ import annotations
import json
import os
from typing import Callable
import asyncio
import inspect
from io import TextIOWrapper
import pathlib

from pydantic import BaseModel

from .Logger import logger
from .MusicBots.AbstractMusicBot import AbstractMusicBotConfig
from .MusicBots.FredBoat import FredboatMusicBotConfig

class KeyBind(BaseModel):
    sequence: list[str] 

class AudioKeyBind(KeyBind):
    audio: str

class StopKeyBind(KeyBind):
    pass

class QuitKeyBind(KeyBind):
    pass

class Config(BaseModel):
    token: str
    stop_keybind: StopKeyBind
    quit_keybind: QuitKeyBind
    keybinds: list[AudioKeyBind] = []
    music_bots: list[AbstractMusicBotConfig] = []

class GeneratedConfigError(RuntimeError):
    pass

class Settings:
    config: Config

    __on_change_callbacks = list[Callable]()

    def __init__(self, path: str, schema_path = "config.schema.json"):
        self.path = path
        self.schema_path = pathlib.Path(schema_path).resolve()

        if not self.schema_path.exists():
            with self.schema_path.open("w") as f:
                logger.info("Writing schema file to " + schema_path)
                self.__write_schema_file(f)

        default_config = Config(
                token="Your Token Here",
                stop_keybind=StopKeyBind(sequence=["control", "shift", "0"]),
                quit_keybind=QuitKeyBind(sequence=["control", "shift", "q"]),
                keybinds=[
                    AudioKeyBind(sequence=["control", "shift", "5"], audio="https://www.youtube.com/watch?v=dQw4w9WgXcQ")
                ],
                music_bots=[
                    FredboatMusicBotConfig(channel_id="Your channel ID here")
                ]
            )

        if os.path.exists(path):
            try:
                self.__read_file()

            except (json.JSONDecodeError, TypeError):
                logger.info("Unable to parse config file. Generating clean config...")

                self.config = default_config
                self.__write_file()
                raise GeneratedConfigError()

        else:
            self.config = default_config
            self.__write_file()
            raise GeneratedConfigError()

    def __write_schema_file(self, file: TextIOWrapper):
        file.write(Config.schema_json(indent=4))

    def __enter__(self):
        self.__start_watching_file()
        return self

    def __exit__(self, *args):
        if self.__is_watching:
            self.__stop_watching_file()

    def on_change(self, func):
        self.__on_change_callbacks.append(func)
        return func

    def __start_watching_file(self):
        self.__is_watching = True
        asyncio.create_task(self.__watch_file())

    def __stop_watching_file(self):
        self.__is_watching = False

    async def __watch_file(self):
        last_change = os.path.getmtime('config.json')

        while self.__is_watching:
            if os.path.getmtime('config.json') != last_change:

                last_change = os.path.getmtime('config.json')
                for callback in self.__on_change_callbacks:
                    self.__read_file()
                    if inspect.iscoroutinefunction(callback):
                        await callback()
                    else:
                        callback()
            
            await asyncio.sleep(0.25)

    def __write_file(self):
        """Write current in-memory config to the file system."""
        try:
            with open(self.path, 'w') as f:
                schema_json = {"$schema": "./" + str(self.schema_path.as_posix())}
                config_json = self.__to_json()
                schema_json.update(config_json)

                json.dump(schema_json, f, indent=4)

        except IOError as error:
            logger.error(error)

    def __read_file(self):
        """Read config from file system into memory."""
        try:
            with open(self.path, 'r') as f:
                self.__from_json(json.load(f))

        except IOError as error:
            logger.error(error)

    def __to_json(self) -> dict:
        """Convert in-memory config to JSON-valid Python object."""
        return self.config.dict()

    def __from_json(self, data: str):
        """Convert JSON-valid Python object to in-memory config."""
        self.config = Config(**data)
