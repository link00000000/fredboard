from __future__ import annotations

import json
import os
import dataclasses
from dataclasses import field

from pydantic.dataclasses import dataclass

from fredboard.Logger import logger

@dataclass
class KeyBind:
    sequence: list[str] 
    audio: str

@dataclass
class Config():
    token: str
    channel_id: str
    command_prefix: str
    stop_keybind: list[str]
    quit_keybind: list[str]
    keybinds: list[KeyBind] = field(default_factory=list)

class GeneratedConfigError(RuntimeError):
    pass

class Settings:
    config: Config

    def __init__(self, path: str):
        self.path = path

        __default_keybind = KeyBind(["control", "shift", "5"], "https://www.youtube.com/your-audio")
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

    def __to_json(self) -> str:
        """Convert in-memory config to JSON-valid Python object."""
        return dataclasses.asdict(self.config)

    def __from_json(self, data: str):
        """Convert JSON-valid Python object to in-memory config."""
        self.config = Config(**data)
