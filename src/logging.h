#pragma once

#include <cstdio>
#include <string>

namespace fretboard::logging {
    enum class log_level
    {
        debug,
        info,
        warning,
        error,
        fatal,
    };

    class logger
    {
    public:
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
void fretboard::logging::logger::log(log_level level, const std::string& format, TArgs ...args)
{
    printf(format.c_str(), args...);
}

template <typename ...TArgs>
void fretboard::logging::logger::debug(const std::string& format, TArgs ...args) {
    log(log_level::debug, format, args...);
}

template <typename ...TArgs>
void fretboard::logging::logger::info(const std::string& format, TArgs ...args) {
    log(log_level::info, format, args...);
}

template <typename ...TArgs>
void fretboard::logging::logger::warning(const std::string& format, TArgs ...args) {
    log(log_level::warning, format, args...);
}

template <typename ...TArgs>
void fretboard::logging::logger::error(const std::string& format, TArgs ...args) {
    log(log_level::error, format, args...);
}

template <typename ...TArgs>
void fretboard::logging::logger::fatal(const std::string& format, TArgs ...args) {
    log(log_level::fatal, format, args...);
}
