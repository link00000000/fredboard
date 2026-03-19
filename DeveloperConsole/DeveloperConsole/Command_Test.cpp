#include "Command.h"

#include <algorithm>
#include <catch2/catch_test_macros.hpp>

TEST_CASE("DeveloperConsole/Register command", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand", "a test command", [](std::span<const std::string> Args){}));
}

TEST_CASE("DeveloperConsole/Cannot register command with space in name", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(!DeveloperConsole::RegisterCommand("Test Command", "a test command that contains spaces", [](std::span<const std::string> Args){}));
}

TEST_CASE("DeveloperConsole/Cannot register same command twice", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand", "a test command", [](std::span<const std::string> Args){}));
    REQUIRE(!DeveloperConsole::RegisterCommand("TestCommand", "a test command", [](std::span<const std::string> Args){}));
}

TEST_CASE("DeveloperConsole/Can find a registered command", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand", "a test command", [](std::span<const std::string> Args){}));
    REQUIRE(DeveloperConsole::FindRegisteredCommandDefinition("TestCommand"));
}

TEST_CASE("DeveloperConsole/Cannot find a command that is not registered", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(!DeveloperConsole::FindRegisteredCommandDefinition("TestCommand"));
}

TEST_CASE("DeveloperConsole/Can get all registered commands", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand1", "a test command", [](std::span<const std::string> Args){}));
    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand2", "a test command", [](std::span<const std::string> Args){}));
    REQUIRE(DeveloperConsole::RegisterCommand("TestCommand3", "a test command", [](std::span<const std::string> Args){}));

    const auto RegisteredCommandDefinitions = DeveloperConsole::GetAllRegisteredCommandDefinitions();
    REQUIRE(RegisteredCommandDefinitions.size() == 3);
    REQUIRE(RegisteredCommandDefinitions[0].Name == "TestCommand1");
    REQUIRE(RegisteredCommandDefinitions[1].Name == "TestCommand2");
    REQUIRE(RegisteredCommandDefinitions[2].Name == "TestCommand3");
}

TEST_CASE("DeveloperConsole/AutoCommand registers and unregisters with RAII", "[developerconsole][command]")
{
    DeveloperConsole::UnregisterAllCommands();

    {
        DeveloperConsole::AutoCommand TestCommand("TestCommand", "a test command", [](std::span<const std::string> Args){});
        REQUIRE(DeveloperConsole::FindRegisteredCommandDefinition("TestCommand"));
    }

    REQUIRE(!DeveloperConsole::FindRegisteredCommandDefinition("TestCommand"));
}
