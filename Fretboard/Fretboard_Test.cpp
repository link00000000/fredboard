#include <catch2/catch_test_macros.hpp>

#include "Fretboard.h"

TEST_CASE("Main")
{
    REQUIRE(Fretboard::Main() == 0);
}
