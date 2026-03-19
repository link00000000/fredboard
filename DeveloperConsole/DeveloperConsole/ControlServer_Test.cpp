#include "ControlServer.h"

#include <mutex>
#include <semaphore>
#include <thread>
#include <catch2/catch_test_macros.hpp>
#include <utility>

class MockTransport final : public DeveloperConsole::Transports::ITransport
{
public:
    explicit MockTransport(
          std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> ListenHandler
        , std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> StopHandler
        , std::function<bool(std::span<std::byte>)> SendDataHandler
    )
        : ListenHandler(std::move(ListenHandler))
        , StopHandler(std::move(StopHandler))
        , SendDataHandler(std::move(SendDataHandler))
    {
    }

    Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError> Listen() override { return ListenHandler(); }
    Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError> Stop() override { return StopHandler(); }

    bool SendData(const std::span<std::byte> Data) override { return SendDataHandler(Data); }

private:
    std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> ListenHandler;
    std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> StopHandler;
    std::function<bool(std::span<std::byte>)> SendDataHandler;
};

TEST_CASE("DeveloperConsole/ControlServer/Start and stop a server", "[developerconsole][controlserver]")
{
    std::binary_semaphore TransportListenSem(0);

    auto TransportListenHandler = [&]() -> Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>
    {
        TransportListenSem.acquire();
        return nullptr;
    };

    auto TransportStopHandler = [&]() -> Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>
    {
        TransportListenSem.release();
        return nullptr;
    };

    auto TransportDataReceivedHandler = [&](const std::span<std::byte> Data) -> bool
    {
        return true;
    };


    auto Transport = std::make_unique<MockTransport>(TransportListenHandler, TransportStopHandler, TransportDataReceivedHandler);
    DeveloperConsole::ControlServer Server(std::move(Transport));

    std::thread ServerThread([&Server] { Server.Listen(); });
    Server.Stop();
    ServerThread.join();
}
