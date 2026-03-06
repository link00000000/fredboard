# Library ---------------------------
find_package(dpp CONFIG REQUIRED)

add_library(fretboard_lib
    STATIC
        Fretboard.cpp
        CommandControl/ControlServer.h
        CommandControl/ControlServer_Generic.h
        CommandControl/ControlServer_Macos.h
        CommandControl/ControlServer_Unix.h
        CommandControl/ControlServer_Windows.h
        CommandControl/ControlServer.cpp
        Error/Error.h
        CommandControl/ControlServer_Windows.cpp
)

target_include_directories(fretboard_lib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(fretboard_lib
    PRIVATE
        dpp::dpp
)

if(WIN32)
    target_compile_definitions(fretboard_lib
        PUBLIC
            PLATFORM_WINDOWS=1
    )
elseif(APPLE)
    target_compile_definitions(fretboard_lib
        PUBLIC
            PLATFORM_MACOS=1
    )
elseif(UNIX)
    target_compile_definitions(fretboard_lib
        PUBLIC
            PLATFORM_UNIX=1
    )
else()
    target_compile_definitions(fretboard_lib
        PUBLIC
            PLATFORM_GENERIC=1
    )
endif()

# TODO: Warnings as errors

# Executable ------------------------

add_executable(fretboard
    main.cpp
)

target_include_directories(fretboard
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(fretboard
    PRIVATE
        fretboard_lib
)

# TODO: Warnings as errors

# Tests -----------------------------

find_package(Catch2 3 REQUIRED)

add_executable(fretboard_tests
    Fretboard_Test.cpp
)

target_include_directories(fretboard_tests
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(fretboard_tests
    PRIVATE
        fretboard_lib
        Catch2::Catch2WithMain
)

# TODO: Warnings as errors

include(Catch)
catch_discover_tests(fretboard_tests)
