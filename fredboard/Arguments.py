import argparse

_parser = argparse.ArgumentParser(description="Global hotkeys for Discord music bots")
_parser.add_argument('--color', '-c', dest="enable_colored_output", action="store_true")

arguments = _parser.parse_args()

