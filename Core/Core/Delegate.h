#pragma once

#include <functional>

namespace Core
{
    struct DelegateHandle
    {
        friend struct DelegateHandleGenerator;
        friend struct std::hash<DelegateHandle>;

        bool operator==(const DelegateHandle& Other) const;

        static DelegateHandle Invalid;

    private:
        explicit DelegateHandle(int64_t InValue);

        int64_t Value;
    };

    struct DelegateHandleGenerator
    {
        DelegateHandle GenerateNextHandle();

    private:
        int64_t NextValue = 1;
    };

    template <typename... TArgs>
    struct Delegate
    {
        using TCallback = std::function<void(TArgs...)>;

        DelegateHandle Add(TCallback InHandler)
        {
            DelegateHandle Handle = HandleGenerator.GenerateNextHandle();
            Handlers[Handle] = InHandler;
            return Handle;
        }

        void Remove(DelegateHandle InHandle)
        {
            Handlers.erase(InHandle);
        }

        void Broadcast(TArgs... Args)
        {
            for (auto [_, Callback] : Handlers)
            {
                Callback(Args...);
            }
        }

    private:
        std::unordered_map<DelegateHandle, TCallback> Handlers;
        DelegateHandleGenerator HandleGenerator;
    };
}

template <>
struct std::hash<Core::DelegateHandle>
{
    std::size_t operator()(const Core::DelegateHandle& Handle) const noexcept
    {
        return std::hash<uint64_t>{}(Handle.Value);
    }
};
