from abc import ABC, ABCMeta, abstractmethod

import pydantic
from pydantic import BaseModel

class AbstractMusicBotConifg(BaseModel, ABC):
    name: str
    channel_id: str

    class Config:
        extra = pydantic.Extra.allow

class AbstractMusicBot(metaclass=ABCMeta):
    @abstractmethod
    async def start_audio(self, url: str):
        raise NotImplemented()
    
    @abstractmethod
    async def stop_audio(self):
        raise NotImplemented()

