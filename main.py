import asyncio
import os

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

from fredboard import (DiscordClient, RateLimitError,
        UnauthorizedError, HTTPError, Settings, GeneratedConfigError,
        logger, YoutubeAPI)

is_running = True

def exit():
    global is_running
    is_running = False

def create_bind(client: DiscordClient, audio_url: str, channel_id: str, command_prefix = ";;"):
    async def play_audio():
        try:
            await client.send_message(command_prefix + "play " + audio_url, channel_id)

        except RateLimitError:
            logger.error("Too many requests made too quickly. Try again later.")

        except UnauthorizedError:
            logger.error("Invalid login token. Did you set your login token in config.json?")
            exit()

        except HTTPError as error:
            if error.status == 400:
                logger.error("Bad request. Did you set your channel id in config.json?")

            else:
                raise error

    return play_audio

def create_stop_bind(client: DiscordClient, channel_id: str, command_prefix = ";;"):
    async def stop_audio():
        try:
            await client.send_message(command_prefix + "stop", channel_id)

        except RateLimitError:
            logger.error("Too many requests made too quickly. Try again later.")

        except UnauthorizedError:
            logger.error("Invalid login token. Did you set your login token in config.json?")
            exit()

        except HTTPError as error:
            if error.status == 400:
                logger.error("Bad request. Did you set your channel id in config.json?")

            else:
                raise error

    return stop_audio

async def main():
    settings: Settings
    try:
        settings = Settings("config.json")
    except GeneratedConfigError:
        logger.info("Generated config.json. Update config file before running again.")
        return

    discord = DiscordClient(settings.config.token)
    try:
        logger.info("Connected as {0.username}#{0.discriminator}".format(await discord.id()))
    except UnauthorizedError:
        logger.error("Invalid login token. Did you set your login token in config.json?")
        return

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

    youtube = YoutubeAPI()
    logger.info("Registered global keybinds:")
    logger.info("\t" + "+".join(stop_binding[0]) + " - Stop")
    logger.info("\t" + "+".join(quit_binding[0]) + " - Quit")
    for bind in settings.config.keybinds:
        if youtube.is_youtube_video(bind.audio):
            logger.info("\t" + "+".join(bind.sequence) + ' - YouTube: ' + await youtube.video_title(bind.audio))
        else:
            logger.info("\t" + "+".join(bind.sequence) + ' - ' + bind.audio)

    await youtube.close()

    hotkeys.start_checking_hotkeys()

    while is_running:
        await asyncio.sleep(0.1)

    logger.info("Quitting...")

    hotkeys.stop_checking_hotkeys()
    await discord.close()
    
if __name__ == "__main__":
    asyncio.run(main())
    os.system('pause')

