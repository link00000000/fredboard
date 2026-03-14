#pragma once

#include <string>

#include "Core/Error.h"

namespace DeveloperConsole::Transports
{
    DECLARE_ERROR(Transport, FailedToStart);

    class ITransportListener
    {
        friend class Transport;

    protected:
        virtual ~ITransportListener() = default;

        virtual void OnConnectionOpened() = 0;
        virtual void OnConnectionClosed() = 0;
        virtual void OnMessageReceived(std::string_view Message) = 0;
    };

    class Transport
    {
    public:
        explicit Transport(ITransportListener* InTransportListener);
        virtual ~Transport() = default;

        virtual Core::Result<nullptr_t, TransportError> Listen() = 0;
        virtual Core::Result<nullptr_t, TransportError> Stop() = 0;

    protected:
        virtual void NotifyConnectionOpened();
        virtual void NotifyConnectionClosed();
        virtual auto NotifyMessageReceived(std::string_view Message) -> void;

    private:
        ITransportListener* Listener;
    };
}
