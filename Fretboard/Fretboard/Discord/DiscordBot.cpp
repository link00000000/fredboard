#include "DiscordBot.h"

#include <string>
#include <utility>
#include <dpp/dpp.h>

#include "Core/Log.h"

namespace Fretboard
{
    DiscordBot::DiscordBot(std::string InToken)
        : Token(std::move(InToken))
    {
    }

    void DiscordBot::Run()
    {
        Bot = std::make_unique<dpp::cluster>(Token);

        Bot->on_log([this](const dpp::log_t& Log){ OnLog(Log); });
        Bot->on_slashcommand([this](const dpp::slashcommand_t& Event) { OnSlashCommand(Event); });
        Bot->on_ready([this](const dpp::ready_t& Event) { OnReady(Event); });

        Bot->start(dpp::st_wait);
    }

    void DiscordBot::Stop()
    {
        if (Bot)
        {
            Bot->shutdown();
        }
    }

    void DiscordBot::OnReady(const dpp::ready_t& Event)
    {
        assert(Bot);

        if (dpp::run_once<struct register_bot_commands>())
        {
            Bot->global_command_create(dpp::slashcommand("ping", "Ping pong!", Bot->me.id), [](const dpp::confirmation_callback_t& Confirmation)
            {
                if (Confirmation.is_error())
                {
                    Core::Log::Error("Fretboard::DiscordBot", "Failed to register global command \"Ping\": {}", Confirmation.get_error().message);
                }
            });
        }
    }

    void DiscordBot::OnSlashCommand(const dpp::slashcommand_t& Event)
    {
        assert(Bot.get());

        if (Event.command.get_command_name() == "ping")
        {
            Event.reply("Pong!");
        }
    }

    void DiscordBot::OnLog(const dpp::log_t& Log)
    {
        switch (Log.severity)
        {
        case dpp::ll_trace:
        case dpp::ll_debug:
            Core::Log::Debug("Discord", Log.message);
            break;
        case dpp::ll_info:
            Core::Log::Info("Discord", Log.message);
            break;
        case dpp::ll_warning:
            Core::Log::Warning("Discord", Log.message);
            break;
        case dpp::ll_error:
        case dpp::ll_critical:
            Core::Log::Error("Discord", Log.message);
            break;
        }
    }
}
