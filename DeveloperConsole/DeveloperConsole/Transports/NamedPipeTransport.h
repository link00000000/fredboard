#pragma once

#if PLATFORM_WINDOWS

#include <mutex>
#include <thread>

#include "ITransport.h"
#include "Core/Error.h"

#include <windows.h>

namespace DeveloperConsole::Transports
{
    class NamedPipeTransport final : public ITransport
    {
    public:
        Core::Result<nullptr_t, TransportError> Listen() override;
        Core::Result<nullptr_t, TransportError> Stop() override;

        void SendData(std::span<std::byte> Data) override;
    };
}

#endif // PLATFORM_WINDOWS
