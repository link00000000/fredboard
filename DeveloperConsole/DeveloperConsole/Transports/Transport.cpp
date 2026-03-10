#include "Transport.h"

namespace DeveloperConsole::Transports
{
    Transport::Transport(ITransportListener* InTransportListener)
        : Listener(InTransportListener)
    {
    }

    void Transport::NotifyConnectionOpened()
    {
        if (!Listener)
        {
            return;
        }

        // TODO: Handle threading
        Listener->OnConnectionOpened();
    }

    void Transport::NotifyConnectionClosed()
    {
        if (!Listener)
        {
            return;
        }

        // TODO: Handle threading
        Listener->OnConnectionClosed();
    }

    void Transport::NotifyMessageReceived(const std::string& Message)
    {
        if (!Listener)
        {
            return;
        }

        // TODO: Handle threading
        Listener->OnMessageReceived(Message);
    }
}
