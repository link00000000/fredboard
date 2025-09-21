cmake_minimum_required(VERSION 3.21)
project(fretboard VERSION 0.1 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

file(GLOB_RECURSE SOURCES CONFIGURE_DEPENDS
    src/*.cpp
    src/*.h
)

add_executable(fretboard ${SOURCES})
install(TARGETS fretboard DESTINATION bin)
