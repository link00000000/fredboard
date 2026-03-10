#include "DeveloperConsole/ControlServer.h"

#if PLATFORM_WINDOWS

#include "NamedPipeTransport.h"

#include <iostream>

namespace DeveloperConsole::Transports
{
    constexpr LPCSTR PIPE_NAME   = "\\\\.\\pipe\\MyPipe";
    constexpr DWORD  BUFFER_SIZE = 4096;

    NamedPipeTransport::NamedPipeTransport(ITransportListener* InTransportListener)
        : Transport(InTransportListener)
    {
    }

    Core::Result<nullptr_t, TransportError> NamedPipeTransport::Listen()
    {
        m_running = true;
        m_stopEvent = CreateEventA(nullptr, TRUE, FALSE, nullptr); // TODO: Use CreateEvent macro instead?

        while (m_running)
        {
            HANDLE hPipe = CreateNamedPipeA(
                PIPE_NAME,
                PIPE_ACCESS_DUPLEX,
                PIPE_TYPE_MESSAGE | PIPE_READMODE_MESSAGE | PIPE_WAIT,
                PIPE_UNLIMITED_INSTANCES,
                BUFFER_SIZE, BUFFER_SIZE,
                0, nullptr
            );

            if (hPipe == INVALID_HANDLE_VALUE)
            {
                return TransportError(TransportErrorCode::FailedToStart, std::format("CreateNamedPipe failed with error {}", GetLastError()));
            }

            OVERLAPPED ov      = {};
            ov.hEvent          = CreateEventA(nullptr, TRUE, FALSE, nullptr);
            BOOL connected     = ConnectNamedPipe(hPipe, &ov);
            DWORD err          = GetLastError();

            if (!connected && err == ERROR_IO_PENDING)
            {
                // Wait for either a client connection or a stop signal
                HANDLE waitHandles[] = { ov.hEvent, m_stopEvent };
                DWORD  result = WaitForMultipleObjects(2, waitHandles, FALSE, INFINITE);

                if (result == WAIT_OBJECT_0 + 1)    // stop event fired
                {
                    CloseHandle(ov.hEvent);
                    CloseHandle(hPipe);
                    break;
                }
            }
            else if (!connected && err != ERROR_PIPE_CONNECTED)
            {
                // TODO: Emit error
                // return TransportError(TransportErrorCode::FailedToAcceptIncomingConnection, std::format("ConnectNamedPipe failed with error {}", GetLastError()));
                std::cerr << std::format("ConnectNamedPipe failed with error {}", GetLastError()) << std::endl;

                CloseHandle(ov.hEvent);
                CloseHandle(hPipe);
                continue;
            }

            CloseHandle(ov.hEvent);

            {
                std::lock_guard lock(m_mutex);
                m_clients.push_back(hPipe);
            }

            NotifyConnectionOpened();
            std::thread(&NamedPipeTransport::HandleClient, this, hPipe).detach();
        }

        return nullptr;
    }

    Core::Result<nullptr_t, TransportError> NamedPipeTransport::Stop()
    {
        m_running = false;
        SetEvent(m_stopEvent);  // wakes Listen() and all HandleClient() threads

        for (auto& t : m_clientThreads)
            if (t.joinable()) t.join();

        CloseHandle(m_stopEvent);
        m_stopEvent = INVALID_HANDLE_VALUE;

        std::cout << "Server stopped" << std::endl;
        return nullptr;
    }

    void NamedPipeTransport::HandleClient(HANDLE hPipe)
    {
        char buffer[BUFFER_SIZE];
        OVERLAPPED ov = {};
        ov.hEvent = CreateEventA(nullptr, TRUE, FALSE, nullptr);

        while (m_running)
        {
            ResetEvent(ov.hEvent);

            DWORD bytesRead = 0;
            BOOL ok = ReadFile(hPipe, buffer, sizeof(buffer) - 1, &bytesRead, &ov);
            DWORD err = GetLastError();

            if (!ok && err == ERROR_IO_PENDING)
            {
                // Wait for either data or stop signal
                HANDLE waitHandles[] = {ov.hEvent, m_stopEvent};
                DWORD result = WaitForMultipleObjects(2, waitHandles, FALSE, INFINITE);

                if (result == WAIT_OBJECT_0 + 1) // stop event
                {
                    std::cout << "Client shutting down via stop event" << std::endl;
                    CancelIo(hPipe); // cancel the pending ReadFile
                    break;
                }

                // Read completed — grab the result
                if (!GetOverlappedResult(hPipe, &ov, &bytesRead, FALSE))
                {
                    std::cout << std::format("Client disconnected ({})", GetLastError()) << std::endl;
                    break;
                }
            }
            else if (!ok)
            {
                std::cout << std::format("ReadFile failed: {}", err) << std::endl;
                break;
            }

            buffer[bytesRead] = '\0';
            std::cout << std::format("Received: {}", buffer) << std::endl;

            NotifyMessageReceived(buffer);

            DWORD bytesWritten = 0;
            WriteFile(hPipe, buffer, bytesRead, &bytesWritten, nullptr);
        }

        CloseHandle(ov.hEvent);

        // Remove from client list
        {
            std::lock_guard lock(m_mutex);
            std::erase(m_clients, hPipe);
        }

        DisconnectNamedPipe(hPipe);
        CloseHandle(hPipe);

        std::cout << "Client cleaned up." << std::endl;
        NotifyConnectionClosed();
    }
}

#endif // PLATFORM_WINDOWS
