# Library ---------------------------
find_package(dpp CONFIG REQUIRED)

add_library(Fretboard
    STATIC
        Fretboard/Fretboard.cpp
        Fretboard/Discord/DiscordBot.cpp
)

target_include_directories(Fretboard
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(Fretboard
    PUBLIC
        dpp::dpp
        Core
        DeveloperConsole
)

# TODO: Warnings as errors

# Tests -----------------------------

find_package(Catch2 3 REQUIRED)

add_executable(FretboardTests
        Fretboard/Fretboard_Test.cpp
)

target_include_directories(FretboardTests
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(FretboardTests
    PRIVATE
        Fretboard
        Catch2::Catch2WithMain
)

# TODO: Warnings as errors

include(Catch)
catch_discover_tests(FretboardTests)
