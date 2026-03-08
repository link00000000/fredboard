cmake_minimum_required(VERSION 4.0)

include(cmake/VcpkgBootstrap.cmake)

project(fretboard
    VERSION 1.0.0
    LANGUAGES CXX
)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS)

enable_testing()

add_subdirectory(DeveloperConsole)
add_subdirectory(Core)
add_subdirectory(Fretboard)
add_subdirectory(FretboardApp)
