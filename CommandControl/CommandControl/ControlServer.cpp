#include "CommandControl/ControlServer.h"

#include <optional>

namespace Fretboard::CommandControl
{
    std::optional<Error> ControlServer::RegisterCommand(Command&& InCommand)
    {
        if (InCommand.Name.contains(' '))
        {
            return Error(ErrorCode::InvalidCommandName, std::format("Command name \"{}\" invalid.", InCommand.Name));
        }

        if (Commands.contains(InCommand.Name))
        {
            return Error(ErrorCode::CommandAlreadyRegistered, std::format("Command \"{}\" is already registered.", InCommand.Name));
        }

        Commands[InCommand.Name] = std::move(InCommand);
        return std::nullopt;
    }
}
