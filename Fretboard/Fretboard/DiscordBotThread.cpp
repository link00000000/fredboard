#include "DiscordBotThread.h"

#include <string>

#include "Core/Environment.h"
#include "Core/Log.h"
#include "Discord/DiscordBot.h"

std::unique_ptr<std::thread> Fretboard::StartDiscordBotThread(std::stop_token InStopToken)
{
    std::string DiscordToken = Core::Environment::GetVar("FRETBOARD_DISCORD_TOKEN");

    if (DiscordToken.empty())
    {
        Core::Log::Error("Fretboard::Main", "FRETBOARD_DISCORD_TOKEN not provided");
        return nullptr;
    }

    return std::make_unique<std::thread>([InStopToken, DiscordToken]
    {
        DiscordBot DiscordBot(DiscordToken);
        std::stop_callback(InStopToken, [&DiscordBot]{ DiscordBot.Stop(); });
        DiscordBot.Run();
    });
}
