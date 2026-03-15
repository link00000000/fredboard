#include "CommandRegistry.h"

#include <algorithm>
#include <ranges>
#include <utility>

#include "Core/Log.h"
#include "Core/String.h"

namespace DeveloperConsole
{
    CommandRegistry::CommandRegistry()
        : IntrinsicCommands(CreateIntrinsicCommandContainer(this))
    {
    }

    std::shared_ptr<CommandRegistry> CommandRegistry::GetGlobalRegistry()
    {
        static auto GlobalRegistry = std::make_shared<CommandRegistry>();
        return GlobalRegistry;
    }

    bool CommandRegistry::RegisterCommand(Command InCommand)
    {
        if (InCommand.Name.contains(' '))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command name cannot contain spaces.", InCommand.Name);
            return false;
        }

        if (IntrinsicCommands.Contains(InCommand.Name))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command is a reserved intrinsic command.", InCommand.Name);
            return false;
        }

        if (!RegisteredCommands.Add(InCommand))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command already registered.", InCommand.Name);
            return false;
        }

        return true;
    }

    bool CommandRegistry::UnregisterCommand(const std::string_view CommandName)
    {
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

        return true;
    }

    void CommandRegistry::UnregisterAllCommands()
    {
        RegisteredCommands.Clear();
    }

    std::vector<std::reference_wrapper<const Command>> CommandRegistry::GetAllCommands() const
    {
        std::vector<std::reference_wrapper<const Command>> Commands;

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
        if (const Command* Command = IntrinsicCommands.Find(CommandName))
        {
            Command->Handler(Args);
            return true;
        }

        if (const Command* Command = RegisteredCommands.Find(CommandName))
        {
            Command->Handler(Args);
            return true;
        }

        return false;
    }

    bool CommandRegistry::CommandContainer::Add(Command InCommand)
    {
        auto [_, bSuccess] = Map.emplace(Core::StringUtils::ToLower(InCommand.Name), std::move(InCommand));
        return bSuccess;
    }

    bool CommandRegistry::CommandContainer::Remove(const std::string_view InCommandName)
    {
        return Map.erase(Core::StringUtils::ToLower(InCommandName)) == 1;
    }

    void CommandRegistry::CommandContainer::Clear()
    {
        Map.clear();
    }

    const Command* CommandRegistry::CommandContainer::Find(std::string_view InCommandName) const
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

    bool CommandRegistry::CommandContainer::Contains(const std::string_view InCommandName) const
    {
        return Map.contains(Core::StringUtils::ToLower(InCommandName));
    }

    CommandRegistry::CommandContainer CommandRegistry::CreateIntrinsicCommandContainer(CommandRegistry* Registry)
    {
        CommandContainer Container;

        Container.Add(Command(
            "Help",
            "List available commands",
            [Registry](const std::span<const std::string> Args)
            {
                std::string CommandsList = Registry->GetAllCommands()
                    | std::views::transform([](const Command& InCommand)
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
}

