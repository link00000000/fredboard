#include "Fretboard.h"

#include <iostream>
#include <stop_token>
#include <thread>

#include <dpp/dpp.h>

#include "DeveloperConsoleThread.h"
#include "DiscordBotThread.h"
#include "Core/Log.h"

#include "DeveloperConsole/Transports/NamedPipeTransport.h"

namespace Fretboard
{
    void HandleLog(const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
    {
        std::string LevelName;

        switch (Level)
        {
        case Core::Log::Level::Debug:
            LevelName = "DEBUG";
            break;
        case Core::Log::Level::Info:
            LevelName = "INFO";
            break;
        case Core::Log::Level::Warning:
            LevelName = "WARNING";
            break;
        case Core::Log::Level::Error:
            LevelName = "ERROR";
            break;
        }

        std::println(std::cerr, "[{}] {}: {}", Category, LevelName, Message);
    }

    int Main()
    {
        Core::Log::RegisterOutputHandler(HandleLog);

        std::stop_source StopSource;
        std::vector<std::unique_ptr<std::thread>> Threads;

        if (std::unique_ptr<std::thread> DeveloperConsoleThread = StartDeveloperConsoleThread(StopSource.get_token(), StopSource))
        {
            Threads.emplace_back(std::move(DeveloperConsoleThread));
        }
        else
        {
            StopSource.request_stop();
        }

        if (std::unique_ptr<std::thread> DiscordBotThread = StartDiscordBotThread(StopSource.get_token()))
        {
            if (!StopSource.stop_requested())
            {
                Threads.emplace_back(std::move(DiscordBotThread));
            }
        }
        else
        {
            StopSource.request_stop();
        }

        for (const std::unique_ptr<std::thread>& Thread : Threads)
        {
            assert(Thread);
            Thread->join();
        }

        return 0;
    }
}
