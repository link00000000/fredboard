#include "DeveloperConsole/ControlServer.h"

#include <iostream>

#include "Core/Log.h"
#include "Transports/NamedPipeTransport.h"

namespace DeveloperConsole
{
    ControlServer::~ControlServer()
    {
        CommandRegistry->UnregisterCommand("SendData");
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
        Core::Log::Debug("ControlServer", "Client connected.");
    }

    void ControlServer::OnClientDisconnected()
    {
        Core::Log::Debug("ControlServer", "Client disconnected.");
    }

    void ControlServer::OnDataReceived(std::span<std::byte> Data)
    {
        auto DataStr = std::string_view(reinterpret_cast<const char*>(Data.data()), Data.size());
        Core::Log::Debug("ControlServer", "Received {} bytes of data: \"{}\"", Data.size(), DataStr);

        const auto Tokens = DataStr
            | std::views::split(' ')
            | std::views::transform([](auto Word)
            {
                return std::string(Word.begin(), Word.end());
            })
            | std::ranges::to<std::vector<std::string>>();

        if (Tokens.empty())
        {
            Core::Log::Debug("ControlServer", "Skipping empty received command");
            return;
        }

        const std::string_view Command = Tokens.front();
        std::vector<std::string> Args(
            std::make_move_iterator(Tokens.begin() + 1),
            std::make_move_iterator(Tokens.end())
        );

        if (CommandRegistry->ExecuteOnHandler(Command, Args))
        {
            Core::Log::Debug("ControlServer", "Successfully executed command {}", Command);
        }
        else
        {
            Core::Log::Error("ControlServer", "Failed to execute unknown command {}", Command);
        }

        // TODO
        // 1. Get the command from the string
        // 2. Get the arguments from the string
        // 3. Execute registered command if there is one
        // 4. Output error message if there is no registered command
    }

    bool ControlServer::SendData(const std::span<std::byte> Data)
    {
        Core::Log::Debug("ControlServer", "Sending {} bytes of data.", Data.size());
        return Transport->SendData(Data);
    }
}
