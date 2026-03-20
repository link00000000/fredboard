#pragma once

#if PLATFORM_WINDOWS

#include <mutex>

#include "ITransport.h"

#include <windows.h>

namespace DeveloperConsole::Transports
{
    class NamedPipeTransport final : public ITransport
    {
    public:
        static constexpr DWORD kDefaultBufferSize  = 65536;
        static constexpr DWORD kDefaultReadTimeout = 50; // ms, for overlapped read polling

        explicit NamedPipeTransport(std::string pipeName, DWORD bufferSize = kDefaultBufferSize);
        NamedPipeTransport(const NamedPipeTransport&) = delete; // Non-copyable
        NamedPipeTransport& operator=(const NamedPipeTransport&) = delete; // Non-movable

        ~NamedPipeTransport();

        void Listen() override;
        void Stop() override;

        bool SendData(std::span<std::byte> Data) override;

    private:
        // Creates a new pipe instance ready for ConnectNamedPipe().
        // Returns INVALID_HANDLE_VALUE on failure.
        HANDLE CreatePipeInstance() const;

        // Waits for a single client to connect on hPipe using overlapped I/O
        // so that a stop event can interrupt the wait.
        // Returns true if a client connected, false if stopped.
        bool WaitForClient(HANDLE hPipe);

        // Reads from hPipe in a loop until the client disconnects or Stop() fires.
        void ServiceClient(HANDLE hPipe);

        // Closes a pipe handle and resets it to INVALID_HANDLE_VALUE.
        static void ClosePipe(HANDLE& hPipe);

        std::string         m_pipeName;
        DWORD               m_bufferSize;

        // Written only under m_pipeMutex; read on any thread via Send().
        HANDLE m_hClientPipe{INVALID_HANDLE_VALUE};
        std::mutex m_PipeMutex;

        // Manual-reset event — signalled by Stop(), reset at the top of Listen().
        HANDLE              m_hStopEvent{ nullptr };

        std::atomic<bool>   m_running{ false };
    };
}

#endif // PLATFORM_WINDOWS
