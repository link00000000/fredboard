import argparse

from .Metadata import metadata

_parser = argparse.ArgumentParser(prog=metadata.ProductName, description="Global hotkeys for Discord music bots")
_parser.add_argument('--color', '-c', dest="enable_colored_output", action="store_true")
_parser.add_argument('--version', '-v', action="version", version=metadata.ProductName + ' v' + metadata.Version)

arguments = _parser.parse_args()

