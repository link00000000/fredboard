#include "String.h"

#include <ranges>
#include <locale>

namespace Core::StringUtils
{
    std::string ToLower(std::string_view InStr)
    {
        return InStr
            | std::views::transform([](const char C) { return static_cast<char>(std::tolower(C)); })
            | std::ranges::to<std::string>();
    }
}
