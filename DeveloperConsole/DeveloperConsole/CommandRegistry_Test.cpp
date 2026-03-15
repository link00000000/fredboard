#include "CommandRegistry.h"

#include <algorithm>
#include <catch2/catch_test_macros.hpp>

TEST_CASE("DeveloperConsole/CommandRegistry/Add and execute command in registry", "[developerconsole][commandregistry]")
{
    auto Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
    Registry->UnregisterAllCommands();

    std::vector<std::string> ExecutedArgs;
    const bool bRegistered = Registry->RegisterCommand({
        .Name = "TestCommand",
        .Description = "A test command",
        .Handler = [&ExecutedArgs](std::span<const std::string> Args)
        {
            std::ranges::copy(Args, std::back_inserter(ExecutedArgs));
        },
    });
    REQUIRE(bRegistered);

    REQUIRE(Registry->ExecuteOnHandler("TestCommand", std::vector<std::string>{"One", "Two", "Three"}));

    REQUIRE(std::ranges::equal(ExecutedArgs, std::vector<std::string>{"One", "Two", "Three"}));
}

TEST_CASE("DeveloperConsole/CommandRegistry/Cannot register command with space in name", "[developerconsole][commandregistry]")
{
    auto Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
    Registry->UnregisterAllCommands();

    const bool bRegistered = Registry->RegisterCommand({
        .Name = "Test Command with Spaces",
        .Description = "a test command that contains spaces",
        .Handler = nullptr,
    });

    REQUIRE(!bRegistered);
}

TEST_CASE("DeveloperConsole/CommandRegistry/Cannot register same command twice", "[developerconsole][commandregistry]")
{
    auto Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
    Registry->UnregisterAllCommands();

    REQUIRE(Registry->RegisterCommand({"TestCommand", "A test command", nullptr}));
    REQUIRE(!Registry->RegisterCommand({"TestCommand", "A test command", nullptr}));

    Registry->UnregisterCommand("TestCommand");
    REQUIRE(Registry->RegisterCommand({"TestCommand", "A test command", nullptr}));
}
