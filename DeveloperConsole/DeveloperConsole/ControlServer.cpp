#include "DeveloperConsole/ControlServer.h"

#include <ranges>

#include "Command.h"
#include "Core/Log.h"
#include "Transports/NamedPipeTransport.h"

namespace
{
    constexpr auto g_HelpCommandName = "Help";
    constexpr auto g_SendDataCommandName = "ControlServer.SendData";
}

namespace DeveloperConsole
{
    ControlServer::ControlServer(std::unique_ptr<Transports::ITransport> InTransport)
        : Transport(std::move(InTransport))
    {
        RegisterCommand(g_HelpCommandName, "List available commands", this, &ControlServer::Cmd_Help);
        RegisterCommand(g_SendDataCommandName, "Sends a string as raw data", this, &ControlServer::Cmd_SendData);

        OnClientConnectedDelegateHandle = Transport->OnClientConnectedEvent().Add(this, &ControlServer::OnClientConnected);
        OnClientConnectedDelegateHandle = Transport->OnClientConnectedEvent().Add([this] { OnClientConnected(); });
        OnClientDisconnectedDelegateHandle = Transport->OnClientDisconnectedEvent().Add([this] { OnClientDisconnected(); });
        OnDataReceivedDelegateHandle = Transport->OnDataReceivedEvent().Add([this] (const std::span<std::byte> Data){ OnDataReceived(Data); });
    }

    ControlServer::~ControlServer()
    {
        UnregisterCommand(g_HelpCommandName);
        UnregisterCommand(g_SendDataCommandName);

        Transport->OnClientConnectedEvent().Remove(OnClientConnectedDelegateHandle);
        Transport->OnClientDisconnectedEvent().Remove(OnClientDisconnectedDelegateHandle);
        Transport->OnDataReceivedEvent().Remove(OnDataReceivedDelegateHandle);
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

        const std::string_view CommandName = Tokens.front();

        std::vector<std::string> Args(
            std::make_move_iterator(Tokens.begin() + 1),
            std::make_move_iterator(Tokens.end())
        );

        if (const auto Definition = FindRegisteredCommandDefinition(CommandName))
        {
            Definition->Handler(Args);
            Core::Log::Debug("ControlServer", "Successfully executed command {}", CommandName);
        }
        else
        {
            Core::Log::Error("ControlServer", "Failed to execute unknown command {}", CommandName);
        }
    }

    bool ControlServer::SendData(const std::span<std::byte> Data)
    {
        Core::Log::Debug("ControlServer", "Sending {} bytes of data.", Data.size());
        return Transport->SendData(Data);
    }

    void ControlServer::Cmd_Help(std::span<const std::string> Args)
    {
        std::string CommandsList = GetAllRegisteredCommandDefinitions()
            | std::views::transform([](const CommandDefinition& InCommand)
            {
                return std::format("{}: {}", InCommand.Name, InCommand.Description);
            })
            | std::views::join_with('\n')
            | std::ranges::to<std::string>();

        Core::Log::Info("DeveloperConsole::CommandRegistry", "Registered commands:\n{}", CommandsList);
    }

    void ControlServer::Cmd_SendData(std::span<const std::string> Args)
    {
        auto Bytes = Args
            | std::views::join
            | std::views::transform([](char c) { return static_cast<std::byte>(c); })
            | std::ranges::to<std::vector<std::byte>>();

        SendData(Bytes);
    }
}
