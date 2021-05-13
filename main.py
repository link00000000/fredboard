import asyncio
import os

import aioglobal_hotkeys.aioglobal_hotkeys as hotkeys

from fredboard import (DiscordClient, RateLimitError,
        UnauthorizedError, HTTPError, Settings, GeneratedConfigError,
        logger, YoutubeAPI, FredboatMusicBot, AbstractMusicBot)

is_running = True
shutdown = False

def exit():
    global shutdown
    shutdown = True

def create_bind(music_bot: AbstractMusicBot, url: str):
    async def play_audio():
        await music_bot.start_audio(url)

    return play_audio

def create_stop_bind(music_bot: AbstractMusicBot):
    async def stop_audio():
        await music_bot.stop_audio()

    return stop_audio

async def main():
    settings: Settings
    try:
        settings = Settings("config.json")
    except GeneratedConfigError:
        logger.info("Generated config.json. Update config file before running again.")
        return
    
    @settings.on_change
    def on_settings_file_change():
        logger.info("**Detected config.json change. Reloading config...**")

        global is_running
        is_running = False

    settings.start_watching_file()

    while not shutdown:
        global is_running
        is_running = True

        hotkeys.clear_hotkeys()

        discord = DiscordClient(settings.config.token)
        try:
            logger.info("Connected as {0.username}#{0.discriminator}".format(await discord.id()))
        except UnauthorizedError:
            logger.error("Invalid login token. Did you set your login token in config.json?")
            return

        fredboat = FredboatMusicBot(discord, settings.config.channel_id, settings.config.command_prefix)

        user_bindings = [[
            binding.sequence,
            None,
            create_bind(fredboat, binding.audio)
        ] for binding in settings.config.keybinds]

        stop_binding = [
            settings.config.stop_keybind,
            None,
            create_stop_bind(fredboat)
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

        while is_running and not shutdown:
            await asyncio.sleep(0.1)

        hotkeys.stop_checking_hotkeys()
        await discord.close()

    settings.stop_watching_file()
    
if __name__ == "__main__":
    asyncio.run(main())

    if not shutdown:
        os.system('pause')

