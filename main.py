import asyncio
import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

from fredboard import DiscordClient
from fredboard import Settings

is_running = True

def shutdown():
    global is_running
    is_running = False

async def main():
    settings = Settings("config.json")

    quit_binding = [["control", "shift", "q"], None, shutdown]
    bindings = [[
        binding.sequence,
        lambda: print("Pressed:", binding.audio),
        lambda: print("Released:", binding.audio)
    ] for binding in settings.config.keybinds] + [quit_binding]

    hotkeys.register_hotkeys(bindings)
    hotkeys.start_checking_hotkeys()

    discord = DiscordClient(settings.config.token)
    await discord.send_message("Testing", "112948060086132736")

    while is_running:
        await asyncio.sleep(0.1)

    hotkeys.stop_checking_hotkeys()

if __name__ == "__main__":
    asyncio.run(main())
