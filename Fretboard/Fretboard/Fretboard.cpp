#include "Fretboard.h"

#include <dpp/dpp.h>
#include "CommandControl/ControlServer_Windows.h"

namespace Fretboard
{
    int Main()
    {
#if 0
    	constexpr auto DiscordBotToken = "";
	    dpp::cluster DiscordBot(DiscordBotToken);

	    DiscordBot.on_log(dpp::utility::cout_logger());

	    DiscordBot.on_slashcommand([](const dpp::slashcommand_t& event) {
	        if (event.command.get_command_name() == "ping") {
	            event.reply("Pong!");
	        }
	    });

	    DiscordBot.on_ready([&DiscordBot](const dpp::ready_t& _event) {
			DiscordBot.global_command_create(dpp::slashcommand("ping", "Ping pong!", DiscordBot.me.id));
	    });

	    DiscordBot.start(dpp::st_wait);
#endif

        CommandControl::ControlServerWindows ControlServer;
        ControlServer.Listen();
        return 0;
    }
}
