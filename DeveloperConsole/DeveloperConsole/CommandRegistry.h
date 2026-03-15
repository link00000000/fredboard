#pragma once

#include <complex>
#include <functional>
#include <unordered_set>
#include <span>
#include <string>

#include "Core/String.h"

namespace DeveloperConsole
{
    using CommandHandler = std::function<void(std::span<const std::string>)>;

    struct Command
    {
        std::string Name;
        std::string Description;
        CommandHandler Handler;
    };

    struct CommandRegistry
    {
        CommandRegistry();

        static std::shared_ptr<CommandRegistry> GetGlobalRegistry();

        bool RegisterCommand(Command InCommand);
        bool UnregisterCommand(std::string_view CommandName);
        void UnregisterAllCommands();

        [[nodiscard]] std::vector<std::reference_wrapper<const Command>> GetAllCommands() const;

        [[nodiscard]] bool ExecuteOnHandler(std::string_view CommandName, std::span<const std::string> Args) const;

    private:
        struct CommandContainer
        {
            bool Add(Command InCommand);
            bool Remove(std::string_view InCommandName);
            void Clear();

            [[nodiscard]] const Command* Find(std::string_view InCommandName) const;
            [[nodiscard]] bool Contains(std::string_view InCommandName) const;

            [[nodiscard]] auto begin() const { return Map.begin(); }
            [[nodiscard]] auto end() const { return Map.end(); }

        private:
            std::unordered_map<std::string, Command> Map;
        };

        static CommandContainer CreateIntrinsicCommandContainer(CommandRegistry* Registry);

        // Commands that are built-in to the command registry
        const CommandContainer IntrinsicCommands;

        // Commands that are registered from elsewhere
        CommandContainer RegisteredCommands;
    };
}
