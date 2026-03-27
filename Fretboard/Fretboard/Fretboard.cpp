#include "Fretboard.h"

#include <iostream>
#include <stop_token>
#include <thread>

#include "Core/Log.h"
#include "DeveloperConsole/Command.h"

#include "DeveloperConsole/ControlServer.h"
#include "DeveloperConsole/Transports/NamedPipeTransport.h"
#include <DeveloperConsole/Transports/NullTransport.h>

namespace Fretboard
{
    int Main()
    {
        Core::Log::RegisterOutputHandler([](const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
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
        });

        std::stop_source StopSource;

        DeveloperConsole::RegisterCommand("Shutdown", "Request all subsystems to shutdown and terminate the application.", [StopSource](std::span<const std::string> Args) mutable
        {
            StopSource.request_stop();
        });

        std::thread DeveloperConsoleThread([](const std::stop_token& StopToken)
        {
#if PLATFORM_WINDOWS
            auto Transport = std::make_unique<DeveloperConsole::Transports::NamedPipeTransport>(R"(\\.\pipe\MyDevConsole)");
#else
            auto Transport = std::make_unique<DeveloperConsole::Transports::NullTransport>();
#endif
            DeveloperConsole::ControlServer DeveloperConsole(std::move(Transport));

            std::stop_callback StopCallback(StopToken, [&DeveloperConsole] { DeveloperConsole.Stop(); });

            DeveloperConsole.Listen();
        }, StopSource.get_token());

        DeveloperConsoleThread.join();

        return 0;
    }
}
