# Executable ------------------------

add_executable(FretboardApp
    main.cpp
)

target_link_libraries(FretboardApp
    PUBLIC
        Fretboard
)

# TODO: Warnings as errors
