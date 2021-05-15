import asyncio
import os

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

from fredboard import (DiscordClient, RateLimitError,
        UnauthorizedError, HTTPError, Settings, GeneratedConfigError,
        logger, YoutubeAPI, FredboatMusicBot, AbstractMusicBot)

from fredboard.MusicBots.Types import get_music_bot_type_by_name
from fredboard.BindRegister import BindRegiser

is_running = True
shutdown = False

async def main():
    try:
        with Settings("config.json") as settings:
            @settings.on_change
            def on_settings_file_change():
                logger.info("**Detected config.json change. Reloading config...**")

                global is_running
                is_running = False

            while not shutdown:
                global is_running
                is_running = True

                discord = DiscordClient(settings.config.token)
                try:
                    logger.info("Connected as {0.username}#{0.discriminator}".format(await discord.id()))
                except UnauthorizedError:
                    logger.error("Invalid login token. Did you set your login token in config.json?")
                    return

                music_bots = []
                for bot_config in settings.config.music_bots:
                    try:
                        BotType, BotConfigType = get_music_bot_type_by_name(bot_config.name)
                        music_bots.append(BotType(discord, BotConfigType(**bot_config.dict())))
                    except TypeError as error:
                        logger.error(error)
                
                async with BindRegiser(
                    keybinds=settings.config.keybinds + [settings.config.stop_keybind, settings.config.quit_keybind],
                    music_bots=music_bots
                ) as bind_register:

                    @bind_register.on_quit
                    def on_hotkey_quit():
                        global shutdown
                        shutdown = True

                    while is_running and not shutdown:
                        await asyncio.sleep(0.1)

                await discord.close()

    except GeneratedConfigError:
        logger.info("Generated config.json. Update config file before running again.")
        return
    
if __name__ == "__main__":
    asyncio.run(main())

    if not shutdown:
        os.system('pause')

