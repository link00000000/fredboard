# Library ---------------------------

add_library(fretboard_lib
    STATIC
        Fretboard.cpp
)

target_include_directories(fretboard_lib
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

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

include(Catch)
catch_discover_tests(fretboard_tests)
