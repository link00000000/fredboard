#include "DeveloperConsoleThread.h"

#include "Core/Log.h"
#include "DeveloperConsole/Command.h"
#include "DeveloperConsole/ControlServer.h"
#include "DeveloperConsole/Transports/NamedPipeTransport.h"

std::unique_ptr<std::thread> Fretboard::StartDeveloperConsoleThread(std::stop_token InStopToken, std::stop_source& InStopSource)
{
    DeveloperConsole::RegisterCommand("Shutdown", "Request all subsystems to shutdown and terminate the application.", [InStopSource](std::span<const std::string> Args) mutable
    {
        InStopSource.request_stop();
    });

    return std::make_unique<std::thread>([InStopToken]
    {
        Core::Log::Debug("Fretboard::Main", "Starting DeveloperConsole thread");

#if PLATFORM_WINDOWS
        auto Transport = std::make_unique<DeveloperConsole::Transports::NamedPipeTransport>(R"(\\.\pipe\MyDevConsole)");
#else
        auto Transport = std::make_unique<DeveloperConsole::Transports::NullTransport>();
#endif
        DeveloperConsole::ControlServer DeveloperConsole(std::move(Transport));

        std::stop_callback StopCallback(InStopToken, [&DeveloperConsole] { DeveloperConsole.Stop(); });
        DeveloperConsole.Listen();

        Core::Log::Debug("Fretboard::Main", "Shutting down DeveloperConsole thread");
    });
}
