# Library ---------------------------
add_library(CommandControl
    STATIC
        CommandControl/ControlServer.cpp
        CommandControl/ControlServer_Windows.cpp
)

target_include_directories(CommandControl
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(CommandControl
    PUBLIC
        Core
)

# TODO: Warnings as errors

# Tests -----------------------------

find_package(Catch2 3 REQUIRED)

add_executable(CommandControlTests
        CommandControl/ControlServer_Test.cpp
)

target_include_directories(CommandControlTests
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(CommandControlTests
    PRIVATE
        CommandControl
        Catch2::Catch2WithMain
)

# TODO: Warnings as errors

include(Catch)
catch_discover_tests(CommandControlTests)
