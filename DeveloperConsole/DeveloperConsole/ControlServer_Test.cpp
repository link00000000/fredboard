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
          std::function<void()> ListenHandler
        , std::function<void()> StopHandler
        , std::function<bool(std::span<std::byte>)> SendDataHandler
    )
        : ListenHandler(std::move(ListenHandler))
        , StopHandler(std::move(StopHandler))
        , SendDataHandler(std::move(SendDataHandler))
    {
    }

    void Listen() override { return ListenHandler(); }
    void Stop() override { return StopHandler(); }

    bool SendData(const std::span<std::byte> Data) override { return SendDataHandler(Data); }

private:
    std::function<void()> ListenHandler;
    std::function<void()> StopHandler;
    std::function<bool(std::span<std::byte>)> SendDataHandler;
};

TEST_CASE("DeveloperConsole/ControlServer/Start and stop a server", "[developerconsole][controlserver]")
{
    std::binary_semaphore TransportListenSem(0);

    auto TransportListenHandler = [&]
    {
        TransportListenSem.acquire();
    };

    auto TransportStopHandler = [&]
    {
        TransportListenSem.release();
    };

    auto TransportDataReceivedHandler = [&](const std::span<std::byte> Data)
    {
        return true;
    };

    auto Transport = std::make_unique<MockTransport>(TransportListenHandler, TransportStopHandler, TransportDataReceivedHandler);
    DeveloperConsole::ControlServer Server(std::move(Transport));

    std::thread ServerThread([&Server] { Server.Listen(); });
    Server.Stop();
    ServerThread.join();
}
