#include "Fretboard.h"

#include <iostream>

#include "Core/Log.h"

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

        DeveloperConsole::ControlServer DeveloperConsole(std::in_place_type<DeveloperConsole::Transports::NamedPipeTransport>, R"(\\.\pipe\MyDevConsole)");
        DeveloperConsole.Listen();

        return 0;
    }
}
