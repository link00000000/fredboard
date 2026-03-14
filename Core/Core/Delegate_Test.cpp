#include "Delegate.h"

#include <Catch2/catch_test_macros.hpp>

TEST_CASE("Core/Delegate/Register and broadcast a delegate", "[core][delegate]")
{
    Core::Delegate<const std::string&> TestDelegate;

    std::optional<std::string> BroadcastedValue;
    const auto DelegateHandle = TestDelegate.Add([&](const std::string& Value)
    {
        BroadcastedValue = Value;
    });

    TestDelegate.Broadcast("TestValue");
    TestDelegate.Remove(DelegateHandle);

    REQUIRE((BroadcastedValue.has_value() && BroadcastedValue == "TestValue"));
}

TEST_CASE("Core/Delegate/Register and broadcast a delegate to multiple handlers", "[core][delegate]")
{
    Core::Delegate<const std::string&> TestDelegate;

    std::optional<std::string> BroadcastedValue1;
    const auto DelegateHandle1 = TestDelegate.Add([&](const std::string& Value)
    {
        BroadcastedValue1 = Value;
    });

    std::optional<std::string> BroadcastedValue2;
    const auto DelegateHandle2 = TestDelegate.Add([&](const std::string& Value)
    {
        BroadcastedValue2 = Value;
    });

    TestDelegate.Broadcast("TestValue");
    TestDelegate.Remove(DelegateHandle1);
    TestDelegate.Remove(DelegateHandle2);

    REQUIRE((BroadcastedValue1.has_value() && BroadcastedValue1 == "TestValue"));
    REQUIRE((BroadcastedValue2.has_value() && BroadcastedValue2 == "TestValue"));
}
