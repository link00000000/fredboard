#pragma once

#include <span>
#include <string>

#include "Core/Delegate.h"
#include "Core/Error.h"

namespace DeveloperConsole::Transports
{
    DECLARE_ERROR(Transport, FailedToStart);

    class ITransport
    {
    public:
        virtual ~ITransport() = default;

        virtual Core::Result<nullptr_t, TransportError> Listen() = 0;
        virtual Core::Result<nullptr_t, TransportError> Stop() = 0;

        virtual void SendData(std::span<std::byte> Data) = 0;

        Core::TSDelegate<>& OnClientConnectedEvent() { return OnClientConnectedDelegate; }
        Core::TSDelegate<>& OnClientDisconnectedEvent() { return OnClientDisconnectedDelegate; }
        Core::TSDelegate<std::span<std::byte>>& OnDataReceivedEvent() { return OnDataReceivedDelegate; }

    private:
        Core::TSDelegate<> OnClientConnectedDelegate;
        Core::TSDelegate<> OnClientDisconnectedDelegate;
        Core::TSDelegate<std::span<std::byte>> OnDataReceivedDelegate;
    };
}
