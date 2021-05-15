from abc import ABC, ABCMeta, abstractmethod
from typing import ClassVar

import pydantic
from pydantic import BaseModel

class AbstractMusicBotConfig(BaseModel, ABC):
    id: ClassVar[str]
    name: str
    channel_id: str

    class Config:
        extra = pydantic.Extra.allow

class AbstractMusicBot(metaclass=ABCMeta):
    id: ClassVar[str]

    @abstractmethod
    async def start_audio(self, url: str):
        raise NotImplemented()
    
    @abstractmethod
    async def stop_audio(self):
        raise NotImplemented()

