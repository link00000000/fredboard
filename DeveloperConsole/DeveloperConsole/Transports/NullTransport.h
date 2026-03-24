#pragma once

#include <span>

#include "ITransport.h"

namespace DeveloperConsole::Transports
{
    class NullTransport final : public ITransport
    {
    public:
        void Listen() override {}
        void Stop() override {}

        bool SendData(std::span<std::byte> Data) override { return false; }
    };
}