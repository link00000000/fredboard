#include "CommandRegistry.h"

#include <ranges>
#include <utility>

#include "Core/Log.h"

namespace DeveloperConsole
{
    namespace
    {
        Core::Map<std::string, CommandHandler> CreateIntrinsicCommandHandlerMap(CommandRegistry* Registry)
        {
            Core::Map<std::string, CommandHandler> IntrinsicCommandHandlerMap;

            IntrinsicCommandHandlerMap["help"] = [Registry](const std::span<const std::string> Args)
            {
                std::string CommandsList = Registry->GetAllCommands() | std::views::join_with('\n') | std::ranges::to< std::string>();
                Core::Log::Info("DeveloperConsole::CommandRegistry", "Registered commands:\n{}", CommandsList);
            };

            return IntrinsicCommandHandlerMap;
        }
    }

    CommandRegistry::CommandRegistry()
        : IntrinsicCommandHandlerMap(CreateIntrinsicCommandHandlerMap(this))
    {
    }

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

        if (IntrinsicCommandHandlerMap.contains(Command))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command is a reserved intrinsic command.", Command);
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
        if (IntrinsicCommandHandlerMap.contains(Command))
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove intrinsic command {} from the registry.", Command);
            return;
        }

        if (CommandHandlerMap.erase(Command) == 0)
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove unregistered command {} from the registry.", Command);
        }
    }

    void CommandRegistry::UnregisterAllCommands()
    {
        CommandHandlerMap.clear();
    }

    std::vector<std::string_view> CommandRegistry::GetAllCommands() const
    {
        std::vector<std::string_view> Commands;

        for (const std::string& Command : IntrinsicCommandHandlerMap | std::views::keys)
        {
            Commands.push_back(Command);
        }

        for (const std::string& Command : CommandHandlerMap | std::views::keys)
        {
            Commands.push_back(Command);
        }

        return Commands;
    }

    bool CommandRegistry::ExecuteOnHandler(const std::string_view Command, const std::span<const std::string> Args) const
    {
        if (const auto Handler = IntrinsicCommandHandlerMap.find(Command); Handler != IntrinsicCommandHandlerMap.end())
        {
            Handler->second(Args);
            return true;
        }

        if (const auto Handler = CommandHandlerMap.find(Command); Handler != CommandHandlerMap.end())
        {
            Handler->second(Args);
            return true;
        }

        return false;
    }
}

