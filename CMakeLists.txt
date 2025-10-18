cmake_minimum_required(VERSION 3.21)

project(
    fretboard
    VERSION 0.1
    LANGUAGES CXX
)

# C++ standard
set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

if(CMAKE_CXX_COMPILER_ID MATCHES "Clang|AppleClang|GNU")
    add_compile_options(-fno-inline -fno-inline-small-function -fno-default-inline)
elseif(MSCV)
    add_compile_options(/Ob0)
endif()

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
add_executable(fretboard
    ${SOURCES}
)

target_include_directories(fretboard
    PRIVATE
        ${CMAKE_SOURCE_DIR}/src
        ${DPP_INCLUDE_DIRS}
)

target_link_libraries(fretboard
    PRIVATE
        ${DPP_LIBRARIES}
)

target_compile_options(fretboard
    PRIVATE
        $<$<CXX_COMPILER_ID:GNU,Clang>:-fno-inline -fno-inline-small-functions -fno-default-inline>
        $<$<CXX_COMPILER_ID:MSVC>:/Ob0>
)

# ------------------------------------------------------------------------------
# Installation
# ------------------------------------------------------------------------------
install(TARGETS fretboard
    DESTINATION bin
)
