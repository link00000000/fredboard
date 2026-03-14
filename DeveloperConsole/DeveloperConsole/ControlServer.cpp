#include "DeveloperConsole/ControlServer.h"

#include <iostream>

#include "Transports/NamedPipeTransport.h"

namespace DeveloperConsole
{
    void ControlServer::RegisterCommand(Command&& InCommand)
    {
        if (InCommand.Name.contains(' '))
        {
            // TODO: Handle error
            //return Error(ErrorCode::InvalidCommandName, std::format("Command name \"{}\" invalid.", InCommand.Name));
            return;
        }

        if (Commands.contains(InCommand.Name))
        {
            // TODO: Handle error
            //return Error(ErrorCode::CommandAlreadyRegistered, std::format("Command \"{}\" is already registered.", InCommand.Name));
            return;
        }

        Commands[InCommand.Name] = std::move(InCommand);
    }

    void ControlServer::Listen()
    {
        Transport->Listen();
    }

    void ControlServer::Stop()
    {
        Transport->Stop();
    }

    void ControlServer::OnClientConnected()
    {
        std::cout << "Connection established" << std::endl;
    }

    void ControlServer::OnClientDisconnected()
    {
        std::cout << "Connection closed" << std::endl;
    }

    void ControlServer::OnDataReceived(std::span<std::byte> Data)
    {
        std::cout << "Message received" << std::endl;

        // TODO
        // 1. Get the command from the string
        // 2. Get the arguments from the string
        // 3. Execute registered command if there is one
        // 4. Output error message if there is no registered command
    }
}
