#include "Environment.h"

#include <mutex>

namespace
{
    std::mutex GetEnvVarMutex;
}

std::string Core::Environment::GetVar(const char* Name)
{
    std::lock_guard Lock(GetEnvVarMutex);
    const char* Value = std::getenv(Name);
    if (Value != nullptr)
    {
        return std::string(Value, std::strlen(Value));
    }
    else
    {
        return std::string("");
    }
}
