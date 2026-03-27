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

    // TODO: Make UniqueDelegateHandle that releases with RAII

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

        virtual ~Delegate() = default;

        // TODO: Support passing arbitrary additional args to be forwarded to the handler
        virtual DelegateHandle Add(TCallback InHandler)
        {
            DelegateHandle Handle = HandleGenerator.GenerateNextHandle();
            Handlers[Handle] = InHandler;
            return Handle;
        }

        template <typename T>
        DelegateHandle Add(T* Instance, void (T::*MemberFunc)(TArgs... Args))
        {
            return Add([Instance, MemberFunc](TArgs... Args){ (Instance->*MemberFunc)(std::forward<TArgs>(Args)...); });
        }

        // TODO: Invalidate handle on removal
        // TODO: Support unbinding from the delegate handle: MyDelegateHandle.Unbind();
        virtual void Remove(DelegateHandle InHandle)
        {
            Handlers.erase(InHandle);
        }

        virtual void Broadcast(TArgs... Args) const
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
    struct TSDelegate : public Delegate<TArgs...>
    {
        using Super = Delegate<TArgs...>;
        using TCallback = Super::TCallback;

        using Delegate<TArgs...>::Add;

        virtual DelegateHandle Add(TCallback InHandler) override
        {
            std::lock_guard Lock(Mutex);
            return Super::Add(InHandler);
        }

        virtual void Remove(DelegateHandle Handle) override
        {
            std::lock_guard Lock(Mutex);
            Super::Remove(Handle);
        }

        virtual void Broadcast(TArgs... Args) const override
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
