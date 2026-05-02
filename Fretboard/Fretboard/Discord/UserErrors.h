#pragma once

namespace Discord::UserErrors
{
    inline std::string UnexpectedError()
    {
        static constexpr const char* Error = "An unexpected error occurred.";
        return Error;
    }

    inline std::string CouldNotFindGuild(const dpp::snowflake InGuildId)
    {
        static constexpr const char* Error = "Could not find guild with id {}";
        return std::format(Error, InGuildId.str());
    }

    inline std::string UserNotInVoiceChannel()
    {
        static constexpr const char* Error =
            "You must be in a voice channel on this server to run this command.\n"
            "Try running this command on a server while in a voice channel on that same server.";
        return Error;
    }
}
