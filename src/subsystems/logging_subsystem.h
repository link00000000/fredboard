#pragma once

#include "subsystem.h"
#include <cstdio>
#include <string>

namespace fretboard::subsystems
{

    class logging_subsystem : public fretboard::subsystems::subsystem
    {
    public:
        enum class log_level
        {
            debug,
            info,
            warning,
            error,
            fatal,
        };

        static logging_subsystem& get();

        template <typename ...TArgs>
        void log(log_level level, const std::string& format, TArgs ...args);

        template <typename ...TArgs>
        void debug(const std::string& format, TArgs ...args);

        template <typename ...TArgs>
        void info(const std::string& format, TArgs ...args);

        template <typename ...TArgs>
        void warning(const std::string& format, TArgs ...args);

        template <typename ...TArgs>
        void error(const std::string& format, TArgs ...args);

        template <typename ...TArgs>
        void fatal(const std::string& format, TArgs ...args);
    };
}


template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::log(log_level level, const std::string& format, TArgs ...args)
{
    printf(format.c_str(), args...);
}

template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::debug(const std::string& format, TArgs ...args) {
    log(log_level::debug, format, args...);
}

template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::info(const std::string& format, TArgs ...args) {
    log(log_level::info, format, args...);
}

template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::warning(const std::string& format, TArgs ...args) {
    log(log_level::warning, format, args...);
}

template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::error(const std::string& format, TArgs ...args) {
    log(log_level::error, format, args...);
}

template <typename ...TArgs>
void fretboard::subsystems::logging_subsystem::fatal(const std::string& format, TArgs ...args) {
    log(log_level::fatal, format, args...);
}
