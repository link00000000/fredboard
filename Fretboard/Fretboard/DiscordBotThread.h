#pragma once

#include <memory>
#include <stop_token>

namespace std
{
    class thread;
}

namespace Fretboard
{
    std::unique_ptr<std::thread> StartDiscordBotThread(std::stop_token InStopToken);
}
