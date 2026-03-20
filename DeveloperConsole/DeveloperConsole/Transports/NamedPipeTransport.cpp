#include "DeveloperConsole/ControlServer.h"

#if PLATFORM_WINDOWS

#include "NamedPipeTransport.h"

#include "Core/Log.h"

namespace DeveloperConsole::Transports
{
    NamedPipeTransport::NamedPipeTransport(std::string pipeName, DWORD bufferSize)
        : m_pipeName(std::move(pipeName))
        , m_bufferSize(bufferSize)
    {
        m_hStopEvent = CreateEventW(
            nullptr, // default security
            TRUE, // manual reset – stays signalled until ResetEvent()
            FALSE, // initially non-signalled
            nullptr); // unnamed

        if (!m_hStopEvent)
            throw std::runtime_error("NamedPipeTransport: failed to create stop event");
    }

    NamedPipeTransport::~NamedPipeTransport()
    {
        Stop();

        if (m_hStopEvent)
        {
            CloseHandle(m_hStopEvent);
            m_hStopEvent = nullptr;
        }
    }

    void NamedPipeTransport::Listen()
    {
        // Allow Listen() to be re-entered after a previous Stop().
        m_running.store(true);
        ResetEvent(m_hStopEvent);

        while (m_running.load())
        {
            // --- Create a fresh pipe instance for the next client ---
            HANDLE hPipe = CreatePipeInstance();
            if (hPipe == INVALID_HANDLE_VALUE)
            {
                // If we were stopped while creating the pipe, exit cleanly.
                if (!m_running.load())
                    break;

                // TODO: Print win error
                Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Failed to create pipe instance {}. Trying again in 100ms", m_pipeName);

                // Brief back-off before retrying to avoid a hot spin on repeated
                // system-level failures (e.g. resource exhaustion).
                Sleep(100);
                continue;
            }
            Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Created pipe instance {}", m_pipeName);

            Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Waiting for client to connect");
            // --- Wait for a client to connect (or Stop() to fire) ---
            if (!WaitForClient(hPipe))
            {
                // Stop() was called during the wait.
                ClosePipe(hPipe);
                break;
            }

            // --- Publish the connected pipe so Send() can use it ---
            {
                std::lock_guard lock(m_PipeMutex);
                m_hClientPipe = hPipe;
            }

            Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Client connected");
            OnClientConnectedEvent().Broadcast();

            // --- Pump reads until the client disconnects or Stop() fires ---
            ServiceClient(hPipe);

            Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Client disconnected");
            OnClientDisconnectedEvent().Broadcast();

            // --- Tear down the client pipe ---
            {
                std::lock_guard lock(m_PipeMutex);
                m_hClientPipe = INVALID_HANDLE_VALUE;
            }

            FlushFileBuffers(hPipe);
            DisconnectNamedPipe(hPipe);
            ClosePipe(hPipe);

            Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Pipe closed");

            // Loop back to accept the next client (unless we were stopped).
        }

        // Final cleanup: ensure no pipe handle is left open.
        {
            std::lock_guard lock(m_PipeMutex);
            ClosePipe(m_hClientPipe);
        }

        m_running.store(false);
    }

    void NamedPipeTransport::Stop()
    {
        Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Stop requested");

        m_running.store(false);
        SetEvent(m_hStopEvent); // unblocks both WaitForClient() and ServiceClient()
    }

    bool NamedPipeTransport::SendData(const std::span<std::byte> Data)
    {
        std::lock_guard lock(m_PipeMutex);

        if (Data.empty())
        {
            return false;
        }

        if (m_hClientPipe == INVALID_HANDLE_VALUE)
        {
            return false;
        }

        const auto length = static_cast<DWORD>(Data.size());
        const auto src = reinterpret_cast<const BYTE*>(Data.data());
        DWORD written = 0;
        DWORD total = 0;

        while (total < static_cast<DWORD>(length))
        {
            if (!WriteFile(m_hClientPipe, src + total,
                           static_cast<DWORD>(length) - total,
                           &written, nullptr))
                return false;

            total += written;
        }

        return true;
    }

    HANDLE NamedPipeTransport::CreatePipeInstance() const
    {
        if (m_pipeName.empty())
        {
            return INVALID_HANDLE_VALUE;
        }

        // TODO: Handle errors
        const int32_t PipeNameWideSize = MultiByteToWideChar(
            CP_UTF8,
            0,
            m_pipeName.c_str(),
            static_cast<int32_t>(m_pipeName.size()),
            nullptr,
            0);

        // TODO: Handle errors
        std::wstring PipeNameWide(PipeNameWideSize, L'\0');
        MultiByteToWideChar(
            CP_UTF8,
            0,
            m_pipeName.c_str(),
            static_cast<int32_t>(m_pipeName.size()),
            PipeNameWide.data(),
            PipeNameWideSize);

        return CreateNamedPipeW(
            PipeNameWide.data(),
            PIPE_ACCESS_DUPLEX | // bidirectional
            FILE_FLAG_OVERLAPPED, // required for cancellable waits
            PIPE_TYPE_BYTE |
            PIPE_READMODE_BYTE |
            PIPE_WAIT,
            PIPE_UNLIMITED_INSTANCES, // allow reconnects after disconnect
            m_bufferSize, // outbound buffer
            m_bufferSize, // inbound buffer
            0, // default client timeout
            nullptr); // default security
    }

    bool NamedPipeTransport::WaitForClient(HANDLE hPipe)
    {
        // Use an OVERLAPPED ConnectNamedPipe so we can wait on both the
        // connection and the stop event simultaneously.
        OVERLAPPED ov{};
        ov.hEvent = CreateEventW(nullptr, TRUE, FALSE, nullptr);
        if (!ov.hEvent)
            return false;

        bool clientConnected = false;

        BOOL connected = ConnectNamedPipe(hPipe, &ov);
        if (connected)
        {
            // Synchronous connect (client was already waiting).
            clientConnected = true;
        }
        else
        {
            DWORD err = GetLastError();
            if (err == ERROR_PIPE_CONNECTED)
            {
                // Client connected between CreateNamedPipe and ConnectNamedPipe.
                clientConnected = true;
            }
            else if (err == ERROR_IO_PENDING)
            {
                // Async wait: block until a client arrives OR stop is requested.
                HANDLE waitHandles[2] = {ov.hEvent, m_hStopEvent};
                DWORD result = WaitForMultipleObjects(2, waitHandles, FALSE, INFINITE);

                if (result == WAIT_OBJECT_0)
                {
                    // Connection event fired – confirm success.
                    DWORD transferred = 0;
                    clientConnected = GetOverlappedResult(hPipe, &ov, &transferred, FALSE) != FALSE;
                }
                else
                {
                    // Stop event (or error): cancel the pending connect.
                    CancelIo(hPipe);
                    DWORD transferred = 0;
                    GetOverlappedResult(hPipe, &ov, &transferred, TRUE); // wait for cancel
                    clientConnected = false;
                }
            }
        }

        CloseHandle(ov.hEvent);
        return clientConnected;
    }

    void NamedPipeTransport::ServiceClient(HANDLE hPipe)
    {
        std::vector<std::byte> buffer(m_bufferSize);

        OVERLAPPED ov{};
        ov.hEvent = CreateEventW(nullptr, TRUE, FALSE, nullptr);
        if (!ov.hEvent)
            return;

        while (m_running.load())
        {
            // Issue an async read.
            ResetEvent(ov.hEvent);
            DWORD bytesRead = 0;
            BOOL ok = ReadFile(hPipe, buffer.data(),
                               static_cast<DWORD>(buffer.size()),
                               &bytesRead, &ov);

            if (!ok && GetLastError() != ERROR_IO_PENDING)
            {
                // Client disconnected or fatal error.
                break;
            }

            // Wait for data to arrive OR for Stop() to signal.
            HANDLE waitHandles[2] = {ov.hEvent, m_hStopEvent};
            DWORD result = WaitForMultipleObjects(2, waitHandles, FALSE, INFINITE);

            if (result == WAIT_OBJECT_0)
            {
                // Read completed.
                if (!GetOverlappedResult(hPipe, &ov, &bytesRead, FALSE))
                {
                    // ERROR_BROKEN_PIPE, ERROR_PIPE_NOT_CONNECTED, etc.
                    break;
                }

                if (bytesRead > 0)
                {
                    Core::Log::Debug("DeveloperConsole::NamedPipeTransport", "Received {} bytes of data \"{}\"", bytesRead, std::string(reinterpret_cast<const char*>(buffer.data()), bytesRead));
                    OnDataReceivedEvent().Broadcast(std::span(buffer.begin(), bytesRead));
                }
            }
            else
            {
                // Stop event: cancel the pending read and exit.
                CancelIo(hPipe);
                GetOverlappedResult(hPipe, &ov, &bytesRead, TRUE);
                break;
            }
        }

        CloseHandle(ov.hEvent);
    }

    /*static*/
    void NamedPipeTransport::ClosePipe(HANDLE& hPipe)
    {
        if (hPipe != INVALID_HANDLE_VALUE)
        {
            CloseHandle(hPipe);
            hPipe = INVALID_HANDLE_VALUE;
        }
    }
}

#endif // PLATFORM_WINDOWS
