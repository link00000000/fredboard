#pragma once

#include <string>
#include <unordered_map>

namespace Core
{
    template <typename T, typename U>
    class Map : public std::unordered_map<T, U>
    {
    };

    struct StringHash
    {
        using is_transparent = void; // enables heterogeneous lookup

        size_t operator()(std::string_view sv) const
        {
            return std::hash<std::string_view>{}(sv);
        }
    };

    // Enables lookup using either std::string or std::string_view
    template <typename U>
    class Map<std::string, U> : public std::unordered_map<std::string, U, StringHash, std::equal_to<>>
    {
    };
}
