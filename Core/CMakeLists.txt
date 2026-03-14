# Library ---------------------------
add_library(Core
    STATIC
        Core/Delegate.cpp
        Core/Log.cpp
)

target_include_directories(Core
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

if(WIN32)
    target_compile_definitions(Core
        PUBLIC
            PLATFORM_WINDOWS=1
    )
elseif(APPLE)
    target_compile_definitions(Core
        PUBLIC
            PLATFORM_MACOS=1
    )
elseif(UNIX)
    target_compile_definitions(Core
        PUBLIC
            PLATFORM_UNIX=1
    )
else()
    target_compile_definitions(Core
        PUBLIC
            PLATFORM_GENERIC=1
    )
endif()

target_compile_options(Core PUBLIC $<$<CXX_COMPILER_ID:MSVC>:/Zc:preprocessor>)

# TODO: Warnings as errors

# Tests -----------------------------

find_package(Catch2 3 REQUIRED)

add_executable(CoreTests
        Core/Delegate_Test.cpp
        Core/Log_Test.cpp
)

target_include_directories(CoreTests
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
        $<INSTALL_INTERFACE:include>
)

target_link_libraries(CoreTests
    PRIVATE
        Core
        Catch2::Catch2WithMain
)

# TODO: Warnings as errors

include(Catch)
catch_discover_tests(CoreTests)
