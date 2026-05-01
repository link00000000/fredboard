#pragma once

#include <memory>
#include <string>

#include <dpp/cluster.h>

namespace dpp
{
    struct slashcommand_t;
    struct ready_t;
    struct log_t;
}

namespace Fretboard
{
    class DiscordBot
    {
    public:
        explicit DiscordBot(std::string InToken);

        void Run();
        void Stop();

        void OnReady(const dpp::ready_t& Event);
        void OnSlashCommand(const dpp::slashcommand_t& Event);

        void HandleSlashCommand_Ping(const dpp::slashcommand_t& Event);

    private:
        std::string Token;
        std::unique_ptr<dpp::cluster> Bot;
    };
}
