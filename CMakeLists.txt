cmake_minimum_required(VERSION 3.21)

project(
    fretboard
    VERSION 0.1
    LANGUAGES CXX
)

# C++ standard
set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# ------------------------------------------------------------------------------
# Dependencies
# ------------------------------------------------------------------------------
find_package(PkgConfig REQUIRED)
pkg_check_modules(DPP REQUIRED dpp)

# ------------------------------------------------------------------------------
# Source files
# ------------------------------------------------------------------------------
file(GLOB_RECURSE SOURCES
    CONFIGURE_DEPENDS
    src/*.cpp
    src/*.h
)

# ------------------------------------------------------------------------------
# Executable target
# ------------------------------------------------------------------------------
add_executable(fretboard ${SOURCES})

target_include_directories(fretboard
    PRIVATE
        ${DPP_INCLUDE_DIRS}
)

target_link_libraries(fretboard
    PRIVATE
        ${DPP_LIBRARIES}
)


# ------------------------------------------------------------------------------
# Installation
# ------------------------------------------------------------------------------
install(TARGETS fretboard
    DESTINATION bin
)
