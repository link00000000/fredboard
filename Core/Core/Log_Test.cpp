#include <catch2/catch_test_macros.hpp>

#include "Log.h"

TEST_CASE("Log debug with static log method", "[log]")
{
    std::optional<std::string_view> LoggedCategory;
    std::optional<std::string_view> LoggedMessage;
    std::optional<Core::Log::Level> LoggedLevel;

    Core::Log::RegisterOutputHandler([&](const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
    {
        LoggedCategory = Category;
        LoggedMessage = Message;
        LoggedLevel = Level;
    });

    Core::Log::Debug("TestLogger", "Test");

    REQUIRE((LoggedCategory.has_value() && LoggedCategory == "TestLogger"));
    REQUIRE((LoggedMessage.has_value() && LoggedMessage == "Test"));
    REQUIRE((LoggedLevel.has_value() && LoggedLevel == Core::Log::Level::Debug));
}

TEST_CASE("Log info with static log method", "[log]")
{
    std::optional<std::string_view> LoggedCategory;
    std::optional<std::string_view> LoggedMessage;
    std::optional<Core::Log::Level> LoggedLevel;

    Core::Log::RegisterOutputHandler([&](const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
    {
        LoggedCategory = Category;
        LoggedMessage = Message;
        LoggedLevel = Level;
    });

    Core::Log::Info("TestLogger", "Test");

    REQUIRE((LoggedCategory.has_value() && LoggedCategory == "TestLogger"));
    REQUIRE((LoggedMessage.has_value() && LoggedMessage == "Test"));
    REQUIRE((LoggedLevel.has_value() && LoggedLevel == Core::Log::Level::Info));
}

TEST_CASE("Log warning with static log method", "[log]")
{
    std::optional<std::string_view> LoggedCategory;
    std::optional<std::string_view> LoggedMessage;
    std::optional<Core::Log::Level> LoggedLevel;

    Core::Log::RegisterOutputHandler([&](const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
    {
        LoggedCategory = Category;
        LoggedMessage = Message;
        LoggedLevel = Level;
    });

    Core::Log::Warning("TestLogger", "Test");

    REQUIRE((LoggedCategory.has_value() && LoggedCategory == "TestLogger"));
    REQUIRE((LoggedMessage.has_value() && LoggedMessage == "Test"));
    REQUIRE((LoggedLevel.has_value() && LoggedLevel == Core::Log::Level::Warning));
}

TEST_CASE("Log error with static log method", "[log]")
{
    std::optional<std::string_view> LoggedCategory;
    std::optional<std::string_view> LoggedMessage;
    std::optional<Core::Log::Level> LoggedLevel;

    Core::Log::RegisterOutputHandler([&](const std::string_view Category, const Core::Log::Level Level, const std::string_view Message)
    {
        LoggedCategory = Category;
        LoggedMessage = Message;
        LoggedLevel = Level;
    });

    Core::Log::Error("TestLogger", "Test");

    REQUIRE((LoggedCategory.has_value() && LoggedCategory == "TestLogger"));
    REQUIRE((LoggedMessage.has_value() && LoggedMessage == "Test"));
    REQUIRE((LoggedLevel.has_value() && LoggedLevel == Core::Log::Level::Error));
}

