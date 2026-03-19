#pragma once

#include <functional>
#include <mutex>

#include "Map.h"

namespace Core
{
    struct DelegateHandle
    {
        friend struct DelegateHandleGenerator;
        friend struct std::hash<DelegateHandle>;

        DelegateHandle();

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

        // TODO: Support passing arbitrary additional args to be forwarded to the handler
        DelegateHandle Add(TCallback InHandler)
        {
            DelegateHandle Handle = HandleGenerator.GenerateNextHandle();
            Handlers[Handle] = InHandler;
            return Handle;
        }

        // TODO: Invalidate handle on removal
        // TODO: Support unbinding from the delegate handle: MyDelegateHandle.Unbind();
        void Remove(DelegateHandle InHandle)
        {
            Handlers.erase(InHandle);
        }

        void Broadcast(TArgs... Args) const
        {
            for (auto [_, Callback] : Handlers)
            {
                Callback(Args...);
            }
        }

    private:
        Core::Map<DelegateHandle, TCallback> Handlers;
        DelegateHandleGenerator HandleGenerator;
    };

    /*
     * Thread-safe version of Core::Delegate
     */
    template <typename... TArgs>
    struct TSDelegate : private Delegate<TArgs...>
    {
        using Super = Delegate<TArgs...>;
        using TCallback = Super::TCallback;

        DelegateHandle Add(TCallback InHandler)
        {
            std::lock_guard Lock(Mutex);
            return Super::Add(InHandler);
        }

        void Remove(DelegateHandle Handle)
        {
            std::lock_guard Lock(Mutex);
            Super::Remove(Handle);
        }

        void Broadcast(TArgs... Args) const
        {
            std::lock_guard Lock(Mutex);
            Super::Broadcast(Args...);
        }

    private:
        mutable std::mutex Mutex;
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
