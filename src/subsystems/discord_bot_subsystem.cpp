#include "discord_bot_subsystem.h"
#include "globals.h"

#include <dpp/once.h>
#include <string>

fretboard::subsystems::discord_bot_subsystem& fretboard::subsystems::discord_bot_subsystem::get()
{
    return fretboard::globals::get_subsystem<discord_bot_subsystem>();
}

fretboard::subsystems::discord_bot_subsystem::discord_bot_subsystem(const std::string& token)
    : token(token)
{
    bot = std::make_unique<dpp::cluster>(this->token);

    bot->on_log(dpp::utility::cout_logger());

    bot->on_slashcommand([](const dpp::slashcommand_t& event) {
        if (event.command.get_command_name() == "ping") {
            event.reply("Pong!");
        }
    });

    bot->on_ready([this](const dpp::ready_t& event) {
        if (dpp::run_once<struct register_bot_commands>()) {
            bot->global_command_create(dpp::slashcommand("ping", "Ping pong!", bot->me.id));
        }
    });

    bot->start(dpp::st_return);
}

