#include "Fretboard.h"

#include <iostream>

#include "Core/Log.h"
#include "DeveloperConsole/CommandRegistry.h"

#include "DeveloperConsole/ControlServer.h"
#include "DeveloperConsole/Transports/NamedPipeTransport.h"

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

        std::thread DeveloperConsoleThread([StopSource]() mutable
        {
            std::shared_ptr<DeveloperConsole::CommandRegistry> Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
            Registry->RegisterCommand({
                .Name = "Shutdown",
                .Description = "Request all subsystems to shutdown and terminate the application.",
                .Handler = [StopSource](std::span<const std::string> Args) mutable
                {
                    Core::Log::Debug("Fretboard::Main", "Shutdown requested via command.");
                    StopSource.request_stop();
                },
            });

            DeveloperConsole::ControlServer DeveloperConsole(Registry, std::in_place_type<DeveloperConsole::Transports::NamedPipeTransport>, R"(\\.\pipe\MyDevConsole)");

            std::stop_callback StopCallback(StopSource.get_token(), [&DeveloperConsole]
            {
                DeveloperConsole.Stop();
            });

            DeveloperConsole.Listen();
        });

        DeveloperConsoleThread.join();

        return 0;
    }
}
