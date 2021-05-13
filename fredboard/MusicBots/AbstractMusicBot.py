from abc import ABCMeta, abstractmethod

class AbstractMusicBot(metaclass=ABCMeta):
    def __init__(self, channel_id: str):
        self.channel_id = channel_id

    @abstractmethod
    async def start_audio(self, url: str):
        raise NotImplemented()
    
    @abstractmethod
    async def stop_audio(self):
        raise NotImplemented()

