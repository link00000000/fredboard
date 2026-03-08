#if PLATFORM_WINDOWS
#include "DeveloperConsole/ControlServer_Windows.h"

#include <windows.h>

namespace Fretboard::DeveloperConsole
{
    void ControlServerWindows::Listen()
    {
        while (true)
        {
            HANDLE Pipe = CreateNamedPipeA(
                R"(\\.\pipe\fretboard)",
                PIPE_ACCESS_DUPLEX,
                PIPE_TYPE_BYTE | PIPE_READMODE_BYTE | PIPE_WAIT,
                PIPE_UNLIMITED_INSTANCES,
                4096,
                4096,
                0,
                nullptr
            );

            if (Pipe == INVALID_HANDLE_VALUE)
            {
                std::cerr << "CreateNamedPipeA() failed" << std::endl;
                return;
            }
            else
            {
                std::cerr << "Created named pipe. Waiting for client." << std::endl;
            }

            const bool bConnected = ConnectNamedPipe(Pipe, nullptr) ? true : GetLastError() == ERROR_PIPE_CONNECTED;
            if (bConnected)
            {
                std::cerr << "Client connected." << std::endl;

                char Buffer[1024];
                DWORD BytesRead = 0;

                while (true)
                {
                    BOOL bSuccess = ReadFile(Pipe, Buffer, sizeof(Buffer) - 1, &BytesRead, nullptr);
                    if (!bSuccess || BytesRead == 0)
                    {
                        break;
                    }

                    Buffer[BytesRead] = '\0';

                    std::cout << Buffer << std::endl;

                    // TODO: Execute command here
                    // TODO: Provide ostream that will write to pipe via WriteFile
                }

                FlushFileBuffers(Pipe);
                DisconnectNamedPipe(Pipe);
                CloseHandle(Pipe);
            }
            else
            {
                CloseHandle(Pipe);
            }
        }
    }

    void ControlServerWindows::Stop()
    {
        // TODO: Implement
    }
}
#endif
