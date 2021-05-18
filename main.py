import asyncio
from fredboard.GitHub import GitHub
import os
from signal import signal, SIGINT

from fredboard.Logger import logger
from fredboard.Discord import DiscordClient
from fredboard.Settings import Settings
from fredboard.BindRegister import BindRegiser
from fredboard.BotRegister import BotRegister
from fredboard.Errors import GeneratedConfigError, MalformedConfigError, UnauthorizedError
from fredboard.Metadata import metadata

REPOSITORY = "link00000000/fredboard"

is_running = True
shutdown = False

def exit(*args):
    global shutdown
    shutdown = True

async def main():
    try:
        with Settings("config.json") as settings:
            @settings.on_change
            def on_settings_file_change():
                logger.info("**Detected config.json change. Reloading config...**")

                global is_running
                is_running = False

            async with GitHub(REPOSITORY) as github:
                release = await github.latest_release()
                download_asset = [a for a in release.assets if a.name == 'fredboard.exe'][0]

                current_version = 'v' + metadata.Version
                latest_release_version = release.tag_name

                if current_version != latest_release_version:
                    logger.info(f"Newer version of FredBoard available ({current_version} -> {latest_release_version})")
                    logger.info(f"Download at {download_asset.browser_download_url}")

            while not shutdown:
                global is_running
                is_running = True

                async with DiscordClient(settings.config.token) as discord:
                    try:
                        logger.info("Connected as {0.username}#{0.discriminator}".format(await discord.id()))
                    except UnauthorizedError:
                        logger.error("Invalid login token. Did you set your login token in config.json?")
                        return

                    music_bots, bot_registration_exceptions = await BotRegister.initialize_music_bots_from_config(
                            settings.config.music_bots, discord)

                    for exception in bot_registration_exceptions:
                        logger.error(f"Failed to register bot {exception}")
                    
                    async with BindRegiser(
                        keybinds=settings.config.keybinds + [settings.config.stop_keybind, settings.config.quit_keybind],
                        music_bots=music_bots
                    ) as bind_register:

                        bind_register.on_quit(exit)

                        while is_running and not shutdown:
                            await asyncio.sleep(0.1)

    except GeneratedConfigError:
        logger.info("Generated config.json. Update config file before running again.")
        return

    except MalformedConfigError:
        logger.error("Unable to parse config file. Remove config.json to generate a clean config.")
        return
    
if __name__ == "__main__":
    signal(SIGINT, exit)

    loop = asyncio.get_event_loop()
    loop.run_until_complete(main())

    if not shutdown:
        os.system('pause')

