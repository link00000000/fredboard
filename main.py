import asyncio
import os

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

from fredboard import DiscordClient, RateLimitError, UnauthorizedError
from fredboard import Settings, GeneratedConfigError
from fredboard import logger

is_running = True

def exit():
    global is_running
    is_running = False

def create_bind(client: DiscordClient, audio_url: str, channel_id: str, command_prefix = ";;"):
    async def play_audio():
        try:
            await client.send_message(command_prefix + "play " + audio_url, channel_id)

        except RateLimitError:
            logger.error("Too many requests made to quickly. Try again later.")

        except UnauthorizedError:
            logger.error("Invalid login token. Set your login token in config.json")
            exit()

    return play_audio

def create_stop_bind(client: DiscordClient, channel_id: str, command_prefix = ";;"):
    async def stop_audio():
        try:
            await client.send_message(command_prefix + "stop", channel_id)

        except RateLimitError:
            logger.error("Too many requests made to quickly. Try again later.")

        except UnauthorizedError:
            logger.error("Invalid login token. Set your login token in config.json")
            exit()

    return stop_audio

async def main():
    settings: Settings
    try:
        settings = Settings("config.json")
    except GeneratedConfigError:
        logger.info("Generated config.json. Update config file before running again.")
        return

    discord = DiscordClient(settings.config.token)

    user_bindings = [[
        binding.sequence,
        None,
        create_bind(discord, binding.audio, settings.config.channel_id)
    ] for binding in settings.config.keybinds]

    stop_binding = [
        settings.config.stop_keybind,
        None,
        create_stop_bind(discord, settings.config.channel_id, settings.config.command_prefix)
    ]

    quit_binding = [settings.config.quit_keybind, None, exit]
    

    hotkeys.register_hotkeys(user_bindings + [stop_binding] + [quit_binding])
    hotkeys.start_checking_hotkeys()

    while is_running:
        await asyncio.sleep(0.1)

    hotkeys.stop_checking_hotkeys()
    await discord.close()
    
    os.system('pause')

if __name__ == "__main__":
    asyncio.run(main())
