#pragma once

#include <complex.h>
#include <functional>
#include <span>
#include <string>

#include "Core/Map.h"

namespace DeveloperConsole
{
    using CommandHandler = std::function<void(std::span<const std::string>)>;

    struct CommandRegistry
    {
        static std::shared_ptr<CommandRegistry> GetGlobalRegistry();

        bool RegisterCommand(std::string Command, CommandHandler Handler);
        void UnregisterCommand(std::string Command);
        void UnregisterAllCommands();

        [[nodiscard]] bool ExecuteOnHandler(std::string_view Command, std::span<const std::string> Args) const;

    private:
        Core::Map<std::string, CommandHandler> CommandHandlerMap;
    };

}
