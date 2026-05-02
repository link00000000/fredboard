#include "DiscordBot.h"

#include <string>
#include <utility>
#include <dpp/dpp.h>

#include "UserErrors.h"
#include "LogDiscord.h"
#include "Core/Log.h"

namespace
{
    namespace Command
    {
        const dpp::slashcommand Ping = dpp::slashcommand()
            .set_name("ping")
            .set_description("Ping pong!");

        const dpp::slashcommand Join = dpp::slashcommand()
            .set_name("join")
            .set_description("Join voice channel");

        const dpp::slashcommand Leave = dpp::slashcommand()
            .set_name("leave")
            .set_description("Leave voice channel");

        const dpp::slashcommand Play = dpp::slashcommand()
            .set_name("play")
            .set_description("Play the audio file");
    }

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
    DiscordBot::DiscordBot(std::string InToken) noexcept
        : Token(std::move(InToken))
    {
    }

    void DiscordBot::Run() noexcept
    {
        Bot = std::make_unique<dpp::cluster>(Token);

        Bot->on_log(DppLogHandler{});
        Bot->on_slashcommand([this](const dpp::slashcommand_t& Event) { OnSlashCommand(Event); });
        Bot->on_ready([this](const dpp::ready_t& Event) { OnReady(Event); });

        try
        {
            Bot->start(dpp::st_wait);
        }
        catch (const std::exception& e)
        {
            Core::Log::Error(LogDiscord, "An error occurred while running discord bot: {}", e.what());
        }
    }

    void DiscordBot::Stop() noexcept
    {
        if (Bot)
        {
            try
            {
                Bot->shutdown();
            }
            catch (const std::exception& e)
            {
                Core::Log::Error(LogDiscord, "Failed to stop discord bot: {}", e.what());
            }
        }
    }

    void DiscordBot::OnReady(const dpp::ready_t& InEvent) noexcept
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

        RegisterGlobalCommand(Command::Ping);
        RegisterGlobalCommand(Command::Join);
        RegisterGlobalCommand(Command::Leave);
        RegisterGlobalCommand(Command::Play);
    }

    void DiscordBot::OnSlashCommand(const dpp::slashcommand_t& InEvent) noexcept
    {
        assert(Bot);

        Core::Log::Debug(LogDiscord, "Bot received slash command \"{}\"", InEvent.command.get_command_name());

        if (InEvent.command.get_command_name() == Command::Ping.name)
        {
            const bool bSuccess = HandleSlashCommand_Ping(InEvent);
            if (!bSuccess)
            {
                Core::Log::Error(LogDiscord, "Failed to handle slash command Ping");
            }
        }
        else if (InEvent.command.get_command_name() == Command::Join.name)
        {
            const bool bSuccess = HandleSlashCommand_Join(InEvent);
            if (!bSuccess)
            {
                Core::Log::Error(LogDiscord, "Failed to handle slash command Join");
            }
        }
        else if (InEvent.command.get_command_name() == Command::Leave.name)
        {
            const bool bSuccess = HandleSlashCommand_Leave(InEvent);
            if (!bSuccess)
            {
                Core::Log::Error(LogDiscord, "Failed to handle slash command Leave");
            }
        }
        else if (InEvent.command.get_command_name() == Command::Play.name)
        {
            const bool bSuccess = HandleSlashCommand_Play(InEvent);
            if (!bSuccess)
            {
                Core::Log::Error(LogDiscord, "Failed to handle slash command Play");
            }
        }
        else
        {
            Core::Log::Warning(LogDiscord, "Slash command \"{}\" not handled", InEvent.command.get_command_name());
            NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
        }
    }

    bool DiscordBot::HandleSlashCommand_Ping(const dpp::slashcommand_t& InEvent) noexcept
    {
        try
        {
            InEvent.reply("pong");
            return true;
        }
        catch (const std::exception& e)
        {
            NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
            Core::Log::Error(LogDiscord, "Error while executing slash command Ping: {}", e.what());
            return false;
        }
    }

    bool DiscordBot::HandleSlashCommand_Join(const dpp::slashcommand_t& InEvent) noexcept
    {
        try
        {
            dpp::guild* Guild = dpp::find_guild(InEvent.command.guild_id);
            if (!Guild)
            {
                Core::Log::Error(LogDiscord, "Could not find guild with id {}", InEvent.command.guild_id.str());

                NotifyUserOfError(InEvent, Discord::UserErrors::CouldNotFindGuild(InEvent.command.guild_id));
                return false;
            }

            const auto IssuingUserVoiceChannel = Guild->voice_members.find(InEvent.command.get_issuing_user().id);
            if (IssuingUserVoiceChannel == Guild->voice_members.end())
            {
                Core::Log::Debug(LogDiscord, "User with id {} attempted to issue slash command Join, but is not in a voice channel in guild with id {}",
                    InEvent.command.get_issuing_user().id.str(), Guild->id.str());

                NotifyUserOfError(InEvent, Discord::UserErrors::UserNotInVoiceChannel());
                return true;
            }

            const dpp::voiceconn* CurrentVoiceConnection = InEvent.from()->get_voice(InEvent.command.guild_id);
            const std::optional<dpp::snowflake> CurrentVoiceChannelId = CurrentVoiceConnection ? CurrentVoiceConnection->channel_id : std::optional<dpp::snowflake>{};

            const dpp::voicestate UserVoiceState = IssuingUserVoiceChannel->second;
            const dpp::snowflake UserVoiceChannelId = UserVoiceState.channel_id;

            if (CurrentVoiceChannelId.has_value() && CurrentVoiceChannelId.value() == UserVoiceChannelId)
            {
                Core::Log::Debug(LogDiscord, "Current voice channel (ChannelId = {}) is the same as the issuing user's voice channel id (UserId = {}, ChannelId = {}). Doing nothing.",
                    CurrentVoiceChannelId.has_value() ? CurrentVoiceChannelId.value().str() : "none",
                    InEvent.command.get_issuing_user().id.str(),
                    UserVoiceChannelId.str());

                // Acknowledge the slash command, but don't reply with a message.
                InEvent.reply();

                return true;
            }

            if (CurrentVoiceChannelId.has_value())
            {
                Core::Log::Debug(LogDiscord, "Current voice channel (ChannelId = {}) is not the same as the voice channel of the issuing user (UserId = {}, ChannelId = {}). Disconnecting from current voice channel to prepare to switch.",
                    CurrentVoiceChannelId.has_value() ? CurrentVoiceChannelId.value().str() : "none",
                    InEvent.command.get_issuing_user().id.str(),
                    UserVoiceChannelId.str());

                InEvent.from()->disconnect_voice(InEvent.command.guild_id);
            }

            if (InEvent.owner == nullptr)
            {
                Core::Log::Error(LogDiscord, "Event owner has owner nullptr");

                NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
                return false;
            }
            Core::Log::Debug(LogDiscord, "Joined voice channel with user with id {}", InEvent.command.get_issuing_user().id.str());

            const bool bConnected = Guild->connect_member_voice(*InEvent.owner, InEvent.command.get_issuing_user().id);
            if (!bConnected)
            {
                Core::Log::Error(LogDiscord, "Failed to connect to issuing user's voice channel (UserId = {}, ChannelId = {})",
                    InEvent.command.get_issuing_user().id.str(),
                    UserVoiceChannelId.str());

                NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
                return false;
            }

            // Acknowledge the slash command, but don't reply with a message.
            InEvent.reply();

            return true;
        }
        catch (std::exception& e)
        {
            NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
            Core::Log::Error(LogDiscord, "Error while executing slash command Join: {}", e.what());

            return false;
        }
    }

    bool DiscordBot::HandleSlashCommand_Leave(const dpp::slashcommand_t& InEvent) noexcept
    {
        try
        {
            // TODO: Impl
            return true;
        }
        catch (std::exception& e)
        {
            NotifyUserOfError(InEvent, Discord::UserErrors::UnexpectedError());
            Core::Log::Error(LogDiscord, "Error while executing slash command Leave: {}", e.what());

            return false;
        }
    }

    bool DiscordBot::HandleSlashCommand_Play(const dpp::slashcommand_t& InEvent) noexcept
    {
        // TODO: Try/catch & impl
        return true;
    }

    void DiscordBot::NotifyUserOfError(const dpp::slashcommand_t& InEvent, const std::string& InErrorMessage) noexcept
    {
        const std::string ErrorMessage = !InErrorMessage.empty() ? InErrorMessage : "An unexpected error occurred.";
        try
        {
            InEvent.reply(ErrorMessage);
            Core::Log::Debug(LogDiscord, "Notified user with id {} that the following error occurred: {}", ErrorMessage, InEvent.command.get_issuing_user().id.str());
        }
        catch (std::exception& e)
        {
            Core::Log::Error(LogDiscord, "Failed to notify user with id {} that the following error occurred: {}. Reason = {}", ErrorMessage, InEvent.command.get_issuing_user().id.str(), e.what());
        }
    }
}
