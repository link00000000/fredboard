#pragma once

#include <complex>
#include <functional>
#include <mutex>
#include <unordered_set>
#include <span>
#include <string>

#include "Core/String.h"

namespace DeveloperConsole
{
    using CommandHandler = std::function<void(std::span<const std::string>)>;

    struct CommandDefinition
    {
        std::string Name;
        std::string Description;
        CommandHandler Handler;
    };

    struct CommandRegistry
    {
        CommandRegistry();

        static std::shared_ptr<CommandRegistry> GetGlobalRegistry();

        bool RegisterCommand(CommandDefinition InCommandDefinition);
        bool UnregisterCommand(std::string_view CommandName);
        void UnregisterAllCommands();

        [[nodiscard]] std::vector<CommandDefinition> GetAllCommands() const;

        [[nodiscard]] bool ExecuteOnHandler(std::string_view CommandName, std::span<const std::string> Args) const;

    private:
        struct CommandDefinitionContainer
        {
            bool Add(CommandDefinition InCommandDefinition);
            bool Remove(std::string_view InCommandName);
            void Clear();

            [[nodiscard]] const CommandDefinition* Find(std::string_view InCommandName) const;
            [[nodiscard]] bool Contains(std::string_view InCommandName) const;

            [[nodiscard]] auto begin() const { return Map.begin(); }
            [[nodiscard]] auto end() const { return Map.end(); }

        private:
            std::unordered_map<std::string, CommandDefinition> Map;
        };

        static CommandDefinitionContainer CreateIntrinsicCommandDefinitionContainer(CommandRegistry* Registry);

        // Commands that are built-in to the command registry
        const CommandDefinitionContainer IntrinsicCommands;

        // Commands that are registered from elsewhere
        CommandDefinitionContainer RegisteredCommands;

        mutable std::recursive_mutex Mutex;
    };

    struct ScopedCommand
    {
        ScopedCommand(std::shared_ptr<CommandRegistry> InRegistry, CommandDefinition InCommandDefinition);
        ~ScopedCommand();

    private:
        std::shared_ptr<CommandRegistry> Registry;
        CommandDefinition Command;
    };
}
