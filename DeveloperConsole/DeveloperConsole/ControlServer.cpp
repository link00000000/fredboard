#include "DeveloperConsole/ControlServer.h"

#include "Core/Log.h"
#include "Transports/NamedPipeTransport.h"

namespace DeveloperConsole
{
    ControlServer::~ControlServer()
    {
        CommandRegistry->UnregisterCommand("ControlServer.SendData");
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

        std::vector<std::string> Tokens;
        std::string CurrentToken;

        bool bEscaped = false;
        bool bInQuotes = false;

        for (const char C : DataStr)
        {
            if (bEscaped)
            {
                CurrentToken += C;
                bEscaped = false;
            }
            else if (C == '\\')
            {
                bEscaped = true;
            }
            else if (C == '"')
            {
                bInQuotes = !bInQuotes;
            }
            else if (C == ' ' && !bInQuotes)
            {
                Tokens.push_back(std::move(CurrentToken));
                CurrentToken.clear();
            }
            else
            {
                CurrentToken += C;
            }
        }

        if (!CurrentToken.empty())
        {
            if (!bInQuotes)
            {
                Tokens.push_back(std::move(CurrentToken));
            }
            else
            {
                Core::Log::Error("DeveloperConsole::ControlServer", "Received malformed command ({}). Missing matching terminating quote.", DataStr);
                return;
            }
        }

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
    }

    bool ControlServer::SendData(const std::span<std::byte> Data)
    {
        Core::Log::Debug("ControlServer", "Sending {} bytes of data.", Data.size());
        return Transport->SendData(Data);
    }
}
