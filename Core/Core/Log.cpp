#include "Log.h"

#include <mutex>
#include <utility>

namespace Core::Log
{
    TSDelegate<std::string_view, Level, std::string_view> OnNewLogMessageDelegate;

    DelegateHandle RegisterOutputHandler(OutputHandler Handler)
    {
        return OnNewLogMessageDelegate.Add(std::move(Handler));
    }

    void UnregisterOutputHandler(const DelegateHandle Handle)
    {
        OnNewLogMessageDelegate.Remove(Handle);
    }

    void Debug(const std::string_view Category, const std::string_view Message)
    {
        OnNewLogMessageDelegate.Broadcast(Category, Level::Debug, Message);
    }

    void Info(const std::string_view Category, const std::string_view Message)
    {
        OnNewLogMessageDelegate.Broadcast(Category, Level::Info, Message);
    }

    void Warning(const std::string_view Category, const std::string_view Message)
    {
        OnNewLogMessageDelegate.Broadcast(Category, Level::Warning, Message);
    }

    void Error(const std::string_view Category, const std::string_view Message)
    {
        OnNewLogMessageDelegate.Broadcast(Category, Level::Error, Message);
    }
}

