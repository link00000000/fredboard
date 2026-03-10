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

    class ControlServer final : public Transports::ITransportListener
    {
    public:
        ControlServer();
        virtual ~ControlServer() = default;

        void RegisterCommand(Command&& InCommand);

        void Listen();
        void Stop();

        void OnConnectionOpened() override;
        void OnConnectionClosed() override;
        void OnMessageReceived(const std::string& Message) override;

    private:
        std::unordered_map<std::string, Command> Commands;
        std::unique_ptr<Transports::NamedPipeTransport> Transport;
    };
}
