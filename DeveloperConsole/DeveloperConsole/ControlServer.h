#pragma once

#include "Transports/ITransport.h"

namespace DeveloperConsole
{
    struct CommandRegistry;

    namespace Transports
    {
        class NamedPipeTransport;
    }

    class ControlServer final
    {
    public:
        explicit ControlServer(std::unique_ptr<Transports::ITransport> InTransport);
        ~ControlServer();

        void Listen();
        void Stop();

    private:
        void OnClientConnected();
        void OnClientDisconnected();
        void OnDataReceived(std::span<std::byte> Data);
        bool SendData(std::span<std::byte> Data);

        void Cmd_Help(std::span<const std::string> Args);
        void Cmd_SendData(std::span<const std::string> Args);

        std::unique_ptr<Transports::ITransport> Transport;

        Core::DelegateHandle OnClientConnectedDelegateHandle;
        Core::DelegateHandle OnClientDisconnectedDelegateHandle;
        Core::DelegateHandle OnDataReceivedDelegateHandle;
    };
}
