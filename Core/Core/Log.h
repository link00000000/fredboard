#pragma once

#include <functional>
#include <string_view>

#include "Delegate.h"

namespace Core::Log
{
    enum class Level
    {
        Debug = 0,
        Info = 1,
        Warning = 2,
        Error = 3,
    };

    using OutputHandler = Delegate<std::string_view, Level, std::string_view>::TCallback;

    DelegateHandle RegisterOutputHandler(OutputHandler Handler);
    void UnregisterOutputHandler(DelegateHandle Handle);

    void Debug(std::string_view Category, std::string_view Message);
    void Info(std::string_view Category, std::string_view Message);
    void Warning(std::string_view Category, std::string_view Message);
    void Error(std::string_view Category, std::string_view Message);

    template<typename... TArgs>
    void Debug(const std::string_view Category, const std::format_string<TArgs...> Format, TArgs&&... Args)
    {
        Debug(Category, std::format(Format, std::forward<TArgs>(Args)...));
    }

    template<typename... TArgs>
    void Info(const std::string_view Category, const std::format_string<TArgs...> Format, TArgs&&... Args)
    {
        Info(Category, std::format(Format, std::forward<TArgs>(Args)...));
    }

    template<typename... TArgs>
    void Warning(const std::string_view Category, const std::format_string<TArgs...> Format, TArgs&&... Args)
    {
        Warning(Category, std::format(Format, std::forward<TArgs>(Args)...));
    }

    template<typename... TArgs>
    void Error(const std::string_view Category, const std::format_string<TArgs...> Format, TArgs&&... Args)
    {
        Error(Category, std::format(Format, std::forward<TArgs>(Args)...));
    }
}
