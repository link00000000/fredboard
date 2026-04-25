#include "Command.h"

#include <algorithm>
#include <ranges>
#include <utility>

#include "Core/Log.h"

namespace DeveloperConsole
{
    Core::Map<std::string, CommandDefinition> g_CommandDefinitionMap;
    std::mutex g_CommandDefinitionMapMutex;

    bool RegisterCommand(std::string Name, std::string Description, CommandHandler Handler)
    {
        std::lock_guard Lock(g_CommandDefinitionMapMutex);

        const std::string NormalizedName = Core::String::ToLower(std::move(Name));

        if (NormalizedName.contains(' '))
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command name cannot contain spaces.", Name);
            return false;
        }

        CommandDefinition Definition{std::move(NormalizedName), std::move(Description), std::move(Handler)};
        auto [_, bSuccess] = g_CommandDefinitionMap.emplace(std::make_pair(Definition.Name, Definition));
        if (!bSuccess)
        {
            Core::Log::Error("DeveloperConsole::CommandRegistry", "Failed to register command {}. Command already registered.", Definition.Name);
            return false;
        }

        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Registered command {}.", Definition.Name);
        return true;
    }

    bool UnregisterCommand(std::string_view CommandName)
    {
        std::lock_guard Lock(g_CommandDefinitionMapMutex);

        const std::string NormalizedName = Core::String::ToLower(CommandName);

        if (!g_CommandDefinitionMap.erase(CommandName))
        {
            Core::Log::Warning("DeveloperConsole::CommandRegistry", "Attempted to remove unregistered command {} from the registry.", CommandName);
            return false;
        }

        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Unregistered command {}.", CommandName);
        return true;
    }

    void UnregisterAllCommands()
    {
        std::lock_guard Lock(g_CommandDefinitionMapMutex);

        g_CommandDefinitionMap.clear();
        Core::Log::Debug("DeveloperConsole::CommandRegistry", "Unregistered all commands");
    }

    std::optional<CommandDefinition> FindRegisteredCommandDefinition(std::string_view Name)
    {
        std::lock_guard Lock(g_CommandDefinitionMapMutex);

        const std::string NormalizedName = Core::String::ToLower(Name);

        if (const auto Iter = g_CommandDefinitionMap.find(NormalizedName); Iter != g_CommandDefinitionMap.end())
        {
            Core::Log::Debug("DeveloperConsole::CommandRegistry", "Found command {}", Name);
            return Iter->second;
        }
        else
        {
            Core::Log::Debug("DeveloperConsole::CommandRegistry", "Could not find command {}", Name);
            return std::nullopt;
        }
    }

    std::vector<CommandDefinition> GetAllRegisteredCommandDefinitions()
    {
        std::lock_guard Lock(g_CommandDefinitionMapMutex);

        return g_CommandDefinitionMap
            | std::views::values
            | std::ranges::to<std::vector<CommandDefinition>>();
    }

    AutoCommand::AutoCommand(std::string Name, std::string Description, CommandHandler Handler)
        : CommandName(Name)
    {
        if (!RegisterCommand(std::move(Name), std::move(Description), std::move(Handler)))
        {
            bShouldAutoUnregister = false;
            Core::Log::Error("DeveloperConsole::AutoCommand", "Failed to register auto command {}", CommandName);
        }
    }

    AutoCommand::~AutoCommand()
    {
        if (bShouldAutoUnregister)
        {
            UnregisterCommand(CommandName);
        }
    }
}

