#pragma once

#include <functional>
#include <string>

#include "Core/Error.h"
#include "Transports/ITransport.h"

namespace DeveloperConsole
{
    namespace Transports
    {
        class NamedPipeTransport;
    }

    struct Command
    {
        std::string Name;
        std::function<void(const std::vector<std::string>& Args, std::ostream& OutputDevice)> Handler; // TODO: Use array view?
    };

    class ControlServer
    {
    public:
        template<typename TTransport, typename... TTransportArgs>
        explicit ControlServer(std::in_place_type_t<TTransport>, TTransportArgs&&... args)
            : Transport(std::make_unique<TTransport>(std::forward<TTransportArgs>(args)...))
        {
            Transport->OnClientConnectedEvent().Add([this] { OnClientConnected(); });
            Transport->OnClientDisconnectedEvent().Add([this] { OnClientDisconnected(); });
            Transport->OnDataReceivedEvent().Add([this] (const std::span<std::byte> Data){ OnDataReceived(Data); });
        }

        virtual ~ControlServer() = default;

        void RegisterCommand(Command&& InCommand);

        void Listen();
        void Stop();

    protected:
        virtual void OnClientConnected();
        virtual void OnClientDisconnected();
        virtual void OnDataReceived(std::span<std::byte> Data);

    private:
        std::unordered_map<std::string, Command> Commands;
        std::unique_ptr<Transports::ITransport> Transport;
    };
}
