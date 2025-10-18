#include "main_process.h"

#include <cassert>
#include <chrono>
#include <csignal>
#include <cstring>
#include <thread>

#include "subsystems/discord_bot_subsystem.h"
#include "subsystems/subsystem.h"
#include "globals.h"

namespace
{
    void handle_signal(int signal)
    {
        fretboard::globals::this_process.on_signal(signal);
    }
}

void fretboard::main_process::run()
{
    startup();

    while (!pending_shutdown_request.has_value())
    {
        std::this_thread::sleep_for(std::chrono::milliseconds(100));
    }

    shutdown();
}

void fretboard::main_process::on_signal(int signal)
{
    switch (signal)
    {
        case SIGINT:
        case SIGTERM:
            request_shutdown(fretboard::shutdown_request{ .reason = std::format("received signal {}", strsignal(signal)) });
            break;
        default:
            break;
    }
}

void fretboard::main_process::request_shutdown(const fretboard::shutdown_request& request)
{
    if (!pending_shutdown_request.has_value())
    {
        pending_shutdown_request.emplace(request);
    }
}

fretboard::subsystems::subsystem_collection& fretboard::main_process::get_subsystems()
{
    return subsystems;
}

void fretboard::main_process::startup()
{
    std::signal(SIGINT, handle_signal);

    // TODO: Move token to configuration
    subsystems.regiser_subsystem<fretboard::subsystems::discord_bot_subsystem>("");
}

void fretboard::main_process::shutdown()
{
    subsystems.unregister_all_subsystems();
}

