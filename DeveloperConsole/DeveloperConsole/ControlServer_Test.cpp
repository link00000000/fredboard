#include "ControlServer.h"

#include <mutex>
#include <semaphore>
#include <thread>
#include <catch2/catch_test_macros.hpp>
#include <utility>

class TestTransport final : public DeveloperConsole::Transports::Transport
{
public:
    explicit TestTransport(DeveloperConsole::Transports::ITransportListener* TransportListener
        , std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> ListenHandler
        , std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> StopHandler
    )
        : Transport(TransportListener)
        , ListenHandler(std::move(ListenHandler))
        , StopHandler(std::move(StopHandler))
    {
    }

    Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError> Listen() override { return ListenHandler(); }
    Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError> Stop() override { return StopHandler(); }

private:
    std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> ListenHandler;
    std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> StopHandler;
};

class TestControlServer final : public DeveloperConsole::ControlServer
{
public:
    TestControlServer(
        std::function<void()> OnConnectionOpenedHandler,
        std::function<void()> OnConnectionClosedHandler,
        std::function<void(std::string_view Message)> OnMessageReceivedHandler,
        std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> TransportListenHandler,
        std::function<Core::Result<nullptr_t, DeveloperConsole::Transports::TransportError>()> TransportStopHandler
    )
        : ControlServer(std::in_place_type<TestTransport>, TransportListenHandler, TransportStopHandler)
        , OnConnectionOpenedHandler(std::move(OnConnectionOpenedHandler))
        , OnConnectionClosedHandler(std::move(OnConnectionClosedHandler))
        , OnMessageReceivedHandler(std::move(OnMessageReceivedHandler))
    {
    }

protected:
    void OnConnectionOpened() override { OnConnectionOpenedHandler(); }
    void OnConnectionClosed() override { OnConnectionClosedHandler(); }
    void OnMessageReceived(const std::string_view Message) override { OnMessageReceivedHandler(Message); }

private:
    std::function<void()> OnConnectionOpenedHandler;
    std::function<void()> OnConnectionClosedHandler;
    std::function<void(std::string_view Message)> OnMessageReceivedHandler;
};

TEST_CASE("Start and stop a server", "[developerconsole]")
{
    auto OnConnectionOpened = [](){};
    auto OnConnectionClosed = [](){};
    auto OnMessageReceived = [](std::string_view Message){};

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

    TestControlServer Server(OnConnectionOpened, OnConnectionClosed, OnMessageReceived, TransportListenHandler, TransportStopHandler);

    std::thread ServerThread([&Server]()
    {
        Server.Listen();
    });

    Server.Stop();
    ServerThread.join();
}
