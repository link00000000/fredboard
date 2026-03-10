#pragma once

#if PLATFORM_WINDOWS

#include <mutex>
#include <thread>

#include "Transport.h"
#include "Core/Error.h"

#include <windows.h>

namespace DeveloperConsole::Transports
{
    class NamedPipeTransport final : public Transport
    {
    public:
        explicit NamedPipeTransport(ITransportListener* InTransportListener);

        virtual Core::Result<nullptr_t, TransportError> Listen() override;
        virtual Core::Result<nullptr_t, TransportError> Stop() override;

    private:
        void HandleClient(HANDLE hPipe);

        std::atomic<bool>        m_running   = false;
        HANDLE                   m_stopEvent = INVALID_HANDLE_VALUE;
        std::vector<std::thread> m_clientThreads;
        std::vector<HANDLE>      m_clients;
        std::mutex               m_mutex;  // guards both m_clientThreads and m_clients
    };
}

#endif // PLATFORM_WINDOWS
