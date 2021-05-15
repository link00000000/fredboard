from __future__ import annotations
import json
import os
from typing import Callable
import asyncio
import inspect

from pydantic import BaseModel

from fredboard.Logger import logger

class KeyBind(BaseModel):
    sequence: list[str] 
    audio: str

class Config(BaseModel):
    token: str
    channel_id: str
    command_prefix: str
    stop_keybind: list[str]
    quit_keybind: list[str]
    keybinds: list[KeyBind] = []

class GeneratedConfigError(RuntimeError):
    pass

class Settings:
    config: Config

    __on_change_callbacks = list[Callable]()

    def __init__(self, path: str):
        self.path = path

        __default_keybind = KeyBind(sequence=["control", "shift", "5"], audio="https://www.youtube.com/watch?v=dQw4w9WgXcQ")
        __default_config = Config(
                token="Your Token Here",
                channel_id="Your Channel ID Here",
                command_prefix = ";;",
                stop_keybind=["control", "shift", "0"],
                quit_keybind=["control", "shift", "q"],
                keybinds=[__default_keybind])

        if os.path.exists(path):
            try:
                self.__read_file()

            except (json.JSONDecodeError, TypeError):
                logger.info("Unable to parse config file. Generating clean config...")

                self.config = __default_config
                self.__write_file()
                raise GeneratedConfigError()

        else:
            self.config = __default_config
            self.__write_file()
            raise GeneratedConfigError()

    def on_change(self, func):
        self.__on_change_callbacks.append(func)
        return func

    def start_watching_file(self):
        self.__is_watching = True
        asyncio.create_task(self.__watch_file())

    def stop_watching_file(self):
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
                json.dump(self.__to_json(), f, indent=4)

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
