#pragma once

#include <span>

#include "Core/Delegate.h"

namespace DeveloperConsole::Transports
{
    class ITransport
    {
    public:
        virtual ~ITransport() = default;

        virtual void Listen() = 0;
        virtual void Stop() = 0;

        virtual bool SendData(std::span<std::byte> Data) = 0;

        Core::TSDelegate<>& OnClientConnectedEvent() { return OnClientConnectedDelegate; }
        Core::TSDelegate<>& OnClientDisconnectedEvent() { return OnClientDisconnectedDelegate; }
        Core::TSDelegate<std::span<std::byte>>& OnDataReceivedEvent() { return OnDataReceivedDelegate; }

    private:
        Core::TSDelegate<> OnClientConnectedDelegate;
        Core::TSDelegate<> OnClientDisconnectedDelegate;
        Core::TSDelegate<std::span<std::byte>> OnDataReceivedDelegate;
    };
}
