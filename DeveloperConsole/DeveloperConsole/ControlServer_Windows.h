#pragma once

#if PLATFORM_WINDOWS
#include "ControlServer.h"

namespace Fretboard::DeveloperConsole
{
    class ControlServerWindows final : public ControlServer
    {
    public:
        void Listen() override;
        void Stop() override;
    };
}
#endif
