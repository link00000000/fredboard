#pragma once

#include <functional>
#include <string>

#include "Core/Error.h"
#include "Transports/Transport.h"

namespace DeveloperConsole
{
    namespace Transports
    {
        class NamedPipeTransport;
    }

    struct Command
    {
        std::string Name;
        std::function<void(const std::vector<std::string>& Args, std::ostream& OutputDevice)> Handler; // TODO: Use array view?
    };

    class ControlServer : public Transports::ITransportListener
    {
    public:
        template<typename TTransport, typename... TTransportArgs>
        explicit ControlServer(std::in_place_type_t<TTransport>, TTransportArgs&&... args)
            : Transport(std::make_unique<TTransport>(this, std::forward<TTransportArgs>(args)...))
        {
        }

        virtual ~ControlServer() = default;

        void RegisterCommand(Command&& InCommand);

        void Listen();
        void Stop();

    protected:
        void OnConnectionOpened() override;
        void OnConnectionClosed() override;
        void OnMessageReceived(std::string_view Message) override;

    private:
        std::unordered_map<std::string, Command> Commands;
        std::unique_ptr<Transports::Transport> Transport;
    };
}
