#include "DiscordBot.h"

#include <string>
#include <utility>
#include <dpp/dpp.h>

#include "LogDiscord.h"
#include "Core/Log.h"

namespace
{
    const dpp::slashcommand Cmd_Ping = dpp::slashcommand()
        .set_name("ping")
        .set_description("Ping pong!");

    struct DppLogHandler
    {
        void operator()(const dpp::log_t& Log) const
        {
            switch (Log.severity)
            {
            case dpp::ll_trace:
            case dpp::ll_debug:
                Core::Log::Debug(LogDiscord, Log.message);
                break;
            case dpp::ll_info:
                Core::Log::Info(LogDiscord, Log.message);
                break;
            case dpp::ll_warning:
                Core::Log::Warning(LogDiscord, Log.message);
                break;
            case dpp::ll_error:
            case dpp::ll_critical:
            default:
                Core::Log::Error(LogDiscord, Log.message);
                break;
            }
        }
    };
}

namespace Fretboard
{
    DiscordBot::DiscordBot(std::string InToken)
        : Token(std::move(InToken))
    {
    }

    void DiscordBot::Run()
    {
        Bot = std::make_unique<dpp::cluster>(Token);

        Bot->on_log(DppLogHandler{});
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

        Core::Log::Debug(LogDiscord, "Bot received ready event");

        auto RegisterGlobalCommand = [this](const dpp::slashcommand& InSlashCommand)
        {
            Bot->global_command_create(dpp::slashcommand(InSlashCommand).set_application_id(Bot->me.id), [InSlashCommand](const dpp::confirmation_callback_t& InConfirmation)
            {
                if (InConfirmation.is_error())
                {
                    Core::Log::Error(LogDiscord, "Failed to register global command \"{}\": {}", InSlashCommand.name, InConfirmation.get_error().message);
                }
                else
                {
                    Core::Log::Debug(LogDiscord, "Registered global command \"{}\"", InSlashCommand.name);
                }
            });
        };

        RegisterGlobalCommand(Cmd_Ping);
    }

    void DiscordBot::OnSlashCommand(const dpp::slashcommand_t& Event)
    {
        assert(Bot);

        Core::Log::Debug(LogDiscord, "Bot received slash command \"{}\"", Event.command.get_command_name());

        if (Event.command.get_command_name() == Cmd_Ping.name)
        {
            HandleSlashCommand_Ping(Event);
            return;
        }

        Core::Log::Warning(LogDiscord, "Slash command \"{}\" not handled", Event.command.get_command_name());
    }

    void DiscordBot::HandleSlashCommand_Ping(const dpp::slashcommand_t& Event)
    {
        Event.reply("pong");
    }
}
