#pragma once

#include <utility>

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
        explicit ControlServer(std::shared_ptr<CommandRegistry> CommandRegistry, std::in_place_type_t<TTransport>, TTransportArgs&&... args)
            : CommandRegistry(std::move(CommandRegistry))
            , Transport(std::make_unique<TTransport>(std::forward<TTransportArgs>(args)...))
        {
            Transport->OnClientConnectedEvent().Add([this] { OnClientConnected(); });
            Transport->OnClientDisconnectedEvent().Add([this] { OnClientDisconnected(); });
            Transport->OnDataReceivedEvent().Add([this] (const std::span<std::byte> Data){ OnDataReceived(Data); });
        }

        virtual ~ControlServer() = default;

        void Listen();
        void Stop();

    protected:
        virtual void OnClientConnected();
        virtual void OnClientDisconnected();
        virtual void OnDataReceived(std::span<std::byte> Data);

    private:
        std::shared_ptr<CommandRegistry> CommandRegistry;
        std::unique_ptr<Transports::ITransport> Transport;
    };
}
