#pragma once

#include <string>

#include "Core/Error.h"

namespace DeveloperConsole::Transports
{
    DECLARE_ERROR(Transport, FailedToStart);

    class ITransportListener
    {
    public:
        virtual ~ITransportListener() = default;

        virtual void OnConnectionOpened() = 0;
        virtual void OnConnectionClosed() = 0;
        virtual void OnMessageReceived(const std::string& Message) = 0;
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
        virtual void NotifyMessageReceived(const std::string& Message);

    private:
        ITransportListener* Listener;
    };
}
