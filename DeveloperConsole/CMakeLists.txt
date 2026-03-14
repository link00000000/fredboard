# Library ---------------------------
add_library(DeveloperConsole
    STATIC
        DeveloperConsole/ControlServer.cpp
        DeveloperConsole/Transports/NamedPipeTransport.cpp
)

target_include_directories(DeveloperConsole
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(DeveloperConsole
    PUBLIC
        Core
)

# TODO: Warnings as errors

# Tests -----------------------------

find_package(Catch2 3 REQUIRED)

add_executable(DeveloperConsoleTests
        DeveloperConsole/ControlServer_Test.cpp
)

target_include_directories(DeveloperConsoleTests
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(DeveloperConsoleTests
    PRIVATE
        DeveloperConsole
        Catch2::Catch2WithMain
)

# TODO: Warnings as errors

include(Catch)
catch_discover_tests(DeveloperConsoleTests)
