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
        explicit DiscordBot(std::string InToken) noexcept;

        void Run() noexcept;
        void Stop() noexcept;

        void OnReady(const dpp::ready_t& InEvent) noexcept;
        void OnSlashCommand(const dpp::slashcommand_t& InEvent) noexcept;

        bool HandleSlashCommand_Ping(const dpp::slashcommand_t& InEvent) noexcept;
        bool HandleSlashCommand_Join(const dpp::slashcommand_t& InEvent) noexcept;
        bool HandleSlashCommand_Leave(const dpp::slashcommand_t& InEvent) noexcept;
        bool HandleSlashCommand_Play(const dpp::slashcommand_t& InEvent) noexcept;

    private:
        void NotifyUserOfError(const dpp::slashcommand_t& InEvent, const std::string& InErrorMessage) noexcept;

        std::string Token;
        std::unique_ptr<dpp::cluster> Bot;
    };
}
