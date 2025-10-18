#include "main_process.h"

namespace fretboard::globals
{
    extern main_process this_process;

    template <typename TSubsystem>
    TSubsystem& get_subsystem()
    {
        return this_process.get_subsystems().get_subsystem<TSubsystem>();
    }
}
