#pragma once

#include <memory>
#include <typeindex>
#include <unordered_map>

namespace fretboard::subsystems
{
    class subsystem_collection;

    class subsystem
    {
    };

    class subsystem_collection final
    {
    public:
        template <typename TSubsystem, typename ...TArgs>
        void regiser_subsystem(TArgs ...args);
        void unregister_all_subsystems();

        template <typename TSubsystem>
        TSubsystem& get_subsystem();

    private:
        std::unordered_map<std::type_index, std::unique_ptr<subsystem>> subsystems;
    };
}

template <typename TSubsystem, typename ...TArgs>
void fretboard::subsystems::subsystem_collection::regiser_subsystem(TArgs ...args)
{
    auto new_subsystem = std::make_unique<TSubsystem>(args...);
    subsystems.try_emplace(typeid(*new_subsystem, std::move(new_subsystem)));
}

template <typename TSubsystem>
TSubsystem& fretboard::subsystems::subsystem_collection::get_subsystem()
{
    return *static_cast<TSubsystem*>(subsystems[typeid(TSubsystem)].get());
}
