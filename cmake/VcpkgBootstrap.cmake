set(VCPKG_ROOT "${CMAKE_SOURCE_DIR}/vcpkg")

if (NOT EXISTS "${VCPKG_ROOT}/vcpkg.exe")
    message(STATUS "Bootstrapping vcpkg...")

    if (WIN32)
        execute_process(
                COMMAND bootstrap-vcpkg.bat
                WORKING_DIRECTORY ${VCPKG_ROOT}
                RESULT_VARIABLE result
        )
    else()
        execute_process(
                COMMAND ./bootstrap-vcpkg.sh
                WORKING_DIRECTORY ${VCPKG_ROOT}
                RESULT_VARIABLE result
        )
    endif()

    if (NOT result EQUAL 0)
        message(FATAL_ERROR "vcpkg bootstrap failed")
    endif()
endif()

set(CMAKE_TOOLCHAIN_FILE "${CMAKE_CURRENT_SOURCE_DIR}/vcpkg/scripts/buildsystems/vcpkg.cmake"
        CACHE STRING "Vcpkg toolchain file")