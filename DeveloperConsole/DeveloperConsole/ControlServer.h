#pragma once

#include <ranges>
#include <utility>

#include "CommandRegistry.h"
#include "Core/Error.h"
#include "Transports/ITransport.h"

namespace DeveloperConsole
{
    struct CommandRegistry;

    namespace Transports
    {
        class NamedPipeTransport;
    }

    class ControlServer
    {
    public:
        template<typename TTransport, typename... TTransportArgs>
        explicit ControlServer(std::shared_ptr<CommandRegistry> InCommandRegistry, std::in_place_type_t<TTransport>, TTransportArgs&&... args)
            : CommandRegistry(std::move(InCommandRegistry))
            , Transport(std::make_unique<TTransport>(std::forward<TTransportArgs>(args)...))
        {
            Transport->OnClientConnectedEvent().Add([this] { OnClientConnected(); });
            Transport->OnClientDisconnectedEvent().Add([this] { OnClientDisconnected(); });
            Transport->OnDataReceivedEvent().Add([this] (const std::span<std::byte> Data){ OnDataReceived(Data); });

            CommandRegistry->RegisterCommand("SendData", [this](std::span<const std::string> Args)
            {
                auto Bytes = Args
                    | std::views::join
                    | std::views::transform([](char c) { return static_cast<std::byte>(c); })
                    | std::ranges::to<std::vector<std::byte>>();

                SendData(Bytes);
            });
        }

        virtual ~ControlServer();

        void Listen();
        void Stop();

    protected:
        virtual void OnClientConnected();
        virtual void OnClientDisconnected();
        virtual void OnDataReceived(std::span<std::byte> Data);
        bool SendData(std::span<std::byte> Data);

    private:
        std::shared_ptr<CommandRegistry> CommandRegistry;
        std::unique_ptr<Transports::ITransport> Transport;
    };
}
