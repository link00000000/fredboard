#pragma once

#include <complex>
#include <functional>
#include <span>
#include <string>

#include "Core/Map.h"

namespace DeveloperConsole
{
    using CommandHandler = std::function<void(std::span<const std::string>)>;

    struct CommandRegistry
    {
        CommandRegistry();

        static std::shared_ptr<CommandRegistry> GetGlobalRegistry();

        bool RegisterCommand(std::string Command, CommandHandler Handler);
        void UnregisterCommand(std::string Command);
        void UnregisterAllCommands();

        [[nodiscard]] std::vector<std::string_view> GetAllCommands() const;

        [[nodiscard]] bool ExecuteOnHandler(std::string_view Command, std::span<const std::string> Args) const;

    private:
        // Commands that are built-in to the command registry
        const Core::Map<std::string, CommandHandler> IntrinsicCommandHandlerMap;

        // Commands that are registered from elsewhere
        Core::Map<std::string, CommandHandler> CommandHandlerMap;

    };

}
