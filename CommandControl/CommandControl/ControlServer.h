#pragma once

#include <functional>
#include <iostream>
#include <optional>
#include <string>

#include "Core/Error.h"

namespace Fretboard::CommandControl
{
    enum class ErrorCode
    {
        InvalidCommandName,
        CommandAlreadyRegistered,
    };

    using Error = Error<ErrorCode>;

    struct Command
    {
        std::string Name;
        std::function<void(const std::vector<std::string>& Args, std::ostream& OutputDevice)> Handler; // TODO: Use array view?
    };

    class ControlServer
    {
    public:
        virtual ~ControlServer() = default;

        virtual void Listen() = 0;
        virtual void Stop() = 0;

        std::optional<Error> RegisterCommand(Command&& InCommand);

    protected:
        void HandleClientRequest(std::string ClientRequest)
        {
            std::cout << ClientRequest << std::endl;
            // 1. Get the command from the string
            // 2. Get the arguments from the string
            // 3. Execute registered command if there is one
            // 4. Output error message if there is no registered command
        }

    private:
        std::unordered_map<std::string, Command> Commands;
    };
}
