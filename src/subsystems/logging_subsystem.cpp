#include "logging_subsystem.h"
#include "globals.h"

fretboard::subsystems::logging_subsystem& fretboard::subsystems::logging_subsystem::get()
{
    return fretboard::globals::get_subsystem<logging_subsystem>();
}
