#include "CommandRegistry.h"

#include <utility>

#include "Core/Log.h"

namespace DeveloperConsole
{
    std::shared_ptr<CommandRegistry> CommandRegistry::GetGlobalRegistry()
    {
        static auto GlobalRegistry = std::make_shared<CommandRegistry>();
        return GlobalRegistry;
    }

    bool CommandRegistry::RegisterCommand(std::string Command, CommandHandler Handler)
    {
        if (Command.contains(' '))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command name cannot contain spaces.", Command);
            return false;
        }

        if (CommandHandlerMap.contains(Command))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command already registered.", Command);
            return false;
        }

        CommandHandlerMap[Command] = std::move(Handler);

        return true;
    }

    void CommandRegistry::UnregisterCommand(std::string Command)
    {
        if (CommandHandlerMap.erase(Command) == 0)
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove unregistered command {} from the registry.", Command);
        }
    }

    void CommandRegistry::UnregisterAllCommands()
    {
        CommandHandlerMap.clear();
    }

    bool CommandRegistry::ExecuteOnHandler(const std::string_view Command, const std::span<const std::string> Args) const
    {
        if (const auto Handler = CommandHandlerMap.find(Command); Handler != CommandHandlerMap.end())
        {
            Handler->second(Args);
            return true;
        }

        return false;
    }
}

