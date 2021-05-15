from abc import ABC, ABCMeta, abstractmethod

from pydantic import BaseModel

class AbstractMusicBotConifg(BaseModel, ABC):
    name: str
    channel_id: str

class AbstractMusicBot(metaclass=ABCMeta):
    def __init__(self, config: AbstractMusicBotConifg):
        self.config = config

    @abstractmethod
    async def start_audio(self, url: str):
        raise NotImplemented()
    
    @abstractmethod
    async def stop_audio(self):
        raise NotImplemented()

