#include "CommandRegistry.h"

#include <algorithm>
#include <ranges>
#include <utility>

#include "Core/Log.h"
#include "Core/String.h"

namespace DeveloperConsole
{
    CommandRegistry::CommandRegistry()
        : IntrinsicCommands(CreateIntrinsicCommandDefinitionContainer(this))
    {
    }

    std::shared_ptr<CommandRegistry> CommandRegistry::GetGlobalRegistry()
    {
        static auto GlobalRegistry = std::make_shared<CommandRegistry>();
        return GlobalRegistry;
    }

    bool CommandRegistry::RegisterCommand(CommandDefinition InCommandDefinition)
    {
        std::lock_guard Lock(Mutex);

        if (InCommandDefinition.Name.contains(' '))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command name cannot contain spaces.", InCommandDefinition.Name);
            return false;
        }

        if (IntrinsicCommands.Contains(InCommandDefinition.Name))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command is a reserved intrinsic command.", InCommandDefinition.Name);
            return false;
        }

        if (!RegisteredCommands.Add(InCommandDefinition))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command already registered.", InCommandDefinition.Name);
            return false;
        }

        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Registered command {}.", InCommandDefinition.Name);
        return true;
    }

    bool CommandRegistry::UnregisterCommand(const std::string_view CommandName)
    {
        std::lock_guard Lock(Mutex);

        if (IntrinsicCommands.Contains(CommandName))
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove intrinsic command {} from the registry.", CommandName);
            return false;
        }

        if (!RegisteredCommands.Remove(CommandName))
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove unregistered command {} from the registry.", CommandName);
            return false;
        }

        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Unregistered command {}.", CommandName);
        return true;
    }

    void CommandRegistry::UnregisterAllCommands()
    {
        std::lock_guard Lock(Mutex);

        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Unregistered all commands");
        RegisteredCommands.Clear();
    }

    std::vector<CommandDefinition> CommandRegistry::GetAllCommands() const
    {
        std::lock_guard Lock(Mutex);

        std::vector<CommandDefinition> Commands;

        for (const auto& Command : IntrinsicCommands | std::views::values)
        {
            Commands.push_back(std::cref(Command));
        }

        for (const auto& Command : RegisteredCommands | std::views::values)
        {
            Commands.push_back(std::cref(Command));
        }

        return Commands;
    }

    bool CommandRegistry::ExecuteOnHandler(const std::string_view CommandName, const std::span<const std::string> Args) const
    {
        std::lock_guard Lock(Mutex);

        if (const CommandDefinition* Command = IntrinsicCommands.Find(CommandName))
        {
            Command->Handler(Args);
            return true;
        }

        if (const CommandDefinition* Command = RegisteredCommands.Find(CommandName))
        {
            Command->Handler(Args);
            return true;
        }

        return false;
    }

    bool CommandRegistry::CommandDefinitionContainer::Add(CommandDefinition InCommandDefinition)
    {
        auto [_, bSuccess] = Map.emplace(Core::StringUtils::ToLower(InCommandDefinition.Name), std::move(InCommandDefinition));
        return bSuccess;
    }

    bool CommandRegistry::CommandDefinitionContainer::Remove(const std::string_view InCommandName)
    {
        return Map.erase(Core::StringUtils::ToLower(InCommandName)) == 1;
    }

    void CommandRegistry::CommandDefinitionContainer::Clear()
    {
        Map.clear();
    }

    const CommandDefinition* CommandRegistry::CommandDefinitionContainer::Find(std::string_view InCommandName) const
    {
        auto Iter = std::ranges::find_if(Map, [InCommandName](const auto& Pair)
        {
            return Pair.first == Core::StringUtils::ToLower(InCommandName);
        });

        if (Iter != Map.end())
        {
            return &Iter->second;
        }

        return nullptr;
    }

    bool CommandRegistry::CommandDefinitionContainer::Contains(const std::string_view InCommandName) const
    {
        return Map.contains(Core::StringUtils::ToLower(InCommandName));
    }

    CommandRegistry::CommandDefinitionContainer CommandRegistry::CreateIntrinsicCommandDefinitionContainer(CommandRegistry* Registry)
    {
        CommandDefinitionContainer Container;

        Container.Add(CommandDefinition(
            "Help",
            "List available commands",
            [Registry](const std::span<const std::string> Args)
            {
                std::string CommandsList = Registry->GetAllCommands()
                    | std::views::transform([](const CommandDefinition& InCommand)
                    {
                        return std::format("{}: {}", InCommand.Name, InCommand.Description);
                    })
                    | std::views::join_with('\n')
                    | std::ranges::to<std::string>();

                Core::Log::Info("DeveloperConsole::CommandRegistry", "Registered commands:\n{}", CommandsList);
            }
        ));

        return Container;
    }

    ScopedCommand::ScopedCommand(std::shared_ptr<CommandRegistry> InRegistry, DeveloperConsole::CommandDefinition InCommandDefinition)
        : Registry(std::move(InRegistry))
        , Command(std::move(InCommandDefinition))
    {
        Registry->RegisterCommand(Command);
    }

    ScopedCommand::~ScopedCommand()
    {
        Registry->UnregisterCommand(Command.Name);
    }
}

