#include "CommandRegistry.h"

#include <algorithm>
#include <catch2/catch_test_macros.hpp>

TEST_CASE("Add and execute command in registry", "[developerconsole][commandregistry]")
{
    auto& Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
    Registry.UnregisterAllCommands();

    std::vector<std::string> ExecutedArgs;
    const bool bRegistered = Registry.RegisterCommand("TestCommand", [&ExecutedArgs](std::span<const std::string> Args)
    {
        std::ranges::copy(Args, std::back_inserter(ExecutedArgs));
    });
    REQUIRE(bRegistered);

    Registry.ExecuteOnHandler("TestCommand", std::vector<std::string>{"One", "Two", "Three"});

    REQUIRE(std::ranges::equal(ExecutedArgs, std::vector<std::string>{"One", "Two", "Three"}));
}

TEST_CASE("Cannot register same command twice", "[developerconsole][commandregistry]")
{
    auto& Registry = DeveloperConsole::CommandRegistry::GetGlobalRegistry();
    Registry.UnregisterAllCommands();

    REQUIRE(Registry.RegisterCommand("TestCommand", nullptr));
    REQUIRE(!Registry.RegisterCommand("TestCommand", nullptr));

    Registry.UnregisterCommand("TestCommand");
    REQUIRE(Registry.RegisterCommand("TestCommand", nullptr));
}
