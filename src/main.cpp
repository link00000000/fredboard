#include "globals.h"
#include "main_process.h"

fretboard::main_process fretboard::globals::this_process;

int main()
{
    fretboard::globals::this_process.run();
}
