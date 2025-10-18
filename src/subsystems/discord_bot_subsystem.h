#pragma once

#include <dpp/cluster.h>

#include "subsystem.h"

namespace fretboard::subsystems
{
    class discord_bot_subsystem : public fretboard::subsystems::subsystem
    {
    public:
        static discord_bot_subsystem& get();

        discord_bot_subsystem(const std::string& token);

    private:
        std::string token;
        std::unique_ptr<dpp::cluster> bot;
    };
}
