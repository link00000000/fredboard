#pragma once

#include <complex>
#include <functional>
#include <optional>
#include <span>
#include <string>

#include "Core/String.h"

namespace DeveloperConsole
{
    using CommandHandler = std::function<void(std::span<const std::string> Args)>;

    struct CommandDefinition
    {
        std::string Name;
        std::string Description;
        CommandHandler Handler;
    };

    // TODO: Support passing arbitrary additional args to be forwarded to the handler
    bool RegisterCommand(std::string Name, std::string Description, CommandHandler Handler);

    template<typename T>
    void RegisterCommand(std::string Name, std::string Description, T* Instance, void (T::*MemberFunc)(std::span<const std::string> Args)) {
        RegisterCommand(std::move(Name), std::move(Description), [Instance, MemberFunc](std::span<const std::string> Args){
            (Instance->*MemberFunc)(Args);
        });
    }

    bool UnregisterCommand(std::string_view CommandName);
    void UnregisterAllCommands();
    std::optional<CommandDefinition> FindRegisteredCommandDefinition(std::string_view Name);
    std::vector<CommandDefinition> GetAllRegisteredCommandDefinitions();

    struct AutoCommand final
    {
        AutoCommand(std::string Name, std::string Description, CommandHandler Handler);
        ~AutoCommand();

    private:
        std::string CommandName;
        bool bShouldAutoUnregister = true;
    };
}
