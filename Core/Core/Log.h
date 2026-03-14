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
}
