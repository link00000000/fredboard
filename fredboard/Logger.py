import logging
from logging import FileHandler, LogRecord
from logging.handlers import RotatingFileHandler, TimedRotatingFileHandler
from pathlib import Path
from sys import stdout
import sys
from abc import ABC
from queue import Queue
from threading import Thread

from colorama import Fore, Style

from .Arguments import arguments
from .Utils import is_frozen

class _ColoredFormatter(logging.Formatter):
    COLORS = {
        'WARNING': Fore.YELLOW,
        'INFO': Fore.CYAN,
        'DEBUG': Fore.MAGENTA,
        'CRITICAL': Fore.YELLOW,
        'ERROR': Fore.RED
    }

    def format(self, record: LogRecord):
        level = record.levelname
        record.levelname = self.COLORS[record.levelname] + record.levelname + Style.RESET_ALL

        return logging.Formatter.format(self, record)
        

__formatter = logging.Formatter("[%(asctime)s] %(levelname)s - %(message)s", datefmt="%Y-%m-%d %H:%M:%S")
__colored_formatter = _ColoredFormatter("[%(asctime)s] %(levelname)s - %(message)s", datefmt="%Y-%m-%d %H:%M:%S")

class _AsyncHandler(ABC, object):
    """
    Abstract classes to log files to the disk without blocking the main thread.
    Provides non-blocking file writing of logs. Uses a `Queue` to write to the
    file using a spearate thread.
    @NOTE Should be inheritted from and not used directly
    Adapted from https://github.com/CopterExpress/python-async-logging-handler
    """

    def __init__(self, *args, **kwargs):
        """
        Spawn file logging handler.
        Spawn a file logging handler on a separate thread and estabslish communication
        with the main thread.
        """
        super(_AsyncHandler, self).__init__(*args, **kwargs)
        self.__queue = Queue(-1)
        self.__thread = Thread(target=self.__loop)
        self.__thread.daemon = True
        self.__thread.start()

    def emit(self, record: LogRecord):
        """
        Place new `logging.LogRecord` in logging queue.
        @NOTE Called by Python's built-in logging library. Should *not* be called directly
        Args:
            record (logging.LogRecord): New log message
        """
        record.msg = record.msg.encode('ascii', 'ignore').decode('utf-8')
        self.__queue.put(record)

    def __loop(self):
        """
        Remove `LogRecord`s from the the queue and write them to file.
        @NOTE Should *not* be called directly, this process does not run on the
        main thread.
        """
        while True:
            record = self.__queue.get()
            try:
                super(_AsyncHandler, self).emit(record)
            except:
                pass


class AsyncFileHandler(_AsyncHandler, FileHandler):
    """Non-blocking alternative to `FileHandler`."""

    pass


class AsyncRotatingFileHanlder(_AsyncHandler, RotatingFileHandler):
    """Non-blocking alternative to `RotatingFileHandler`."""

    pass


class AsyncTimedRotatingFileHandler(_AsyncHandler, TimedRotatingFileHandler):
    """Non-blocking alternative to `TimedRotatingFileHandler`."""

    pass

__logger_name = "fredboard"
__log_file_name = __logger_name + ".log"
__log_file_dir = "logs"

logger = logging.getLogger(__logger_name)

log_file_dir = Path(__log_file_dir).resolve()
log_file = Path.resolve(Path.joinpath(log_file_dir, __log_file_name))

# Setup async file logging handler
Path.mkdir(log_file_dir, parents=True, exist_ok=True)
file_handler = AsyncRotatingFileHanlder(log_file)
file_handler.setFormatter(__formatter)
file_handler.setLevel(logging.DEBUG)
logger.addHandler(file_handler)

# Setup stdout logging handler
stdout_handler = logging.StreamHandler(stdout)
stdout_handler.setFormatter(__colored_formatter if arguments.enable_colored_output else __formatter)
stdout_handler.setLevel(logging.DEBUG)
logger.addHandler(stdout_handler)

# Set the logging level of stdout to INFO if compiled to exe
if is_frozen():
    logger.setLevel(logging.INFO)

else:
    logger.setLevel(logging.DEBUG)
