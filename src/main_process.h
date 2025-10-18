#pragma once

#include "subsystems/subsystem.h"
#include <optional>
#include <string>

namespace fretboard
{
    struct shutdown_request
    {
        const std::string reason;
    };

    class main_process
    {
    public:
        void run();
        void on_signal(int signal);
        void request_shutdown(const shutdown_request& request);

        fretboard::subsystems::subsystem_collection& get_subsystems();

    protected:
        void startup();
        void shutdown();

    private:
        fretboard::subsystems::subsystem_collection subsystems;
        std::optional<fretboard::shutdown_request> pending_shutdown_request;
    };
}
