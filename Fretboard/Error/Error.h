#pragma once

#include <memory>
#include <sstream>
#include <string>
#include <type_traits>
#include <utility>

template <typename TErrCode>
requires std::is_enum_v<TErrCode>
struct Error
{
public:
    Error(TErrCode InErrCode, std::string InErrorMessage)
        : ErrorCode(InErrCode), Message(std::move(InErrorMessage)), Cause(nullptr)
    {
    }

    Error(TErrCode InErrCode, std::string InErrorMessage, const Error& InCause)
        : ErrorCode(InErrCode), Message(std::move(InErrorMessage)), Cause(std::make_shared<Error>(std::move(InCause)))
    {
    }

    bool Is(TErrCode InErrCode)
    {
        return ErrorCode == InErrCode;
    }

    const std::string& GetMessage()
    {
        if (Cause)
        {
            return std::format("{}\ncaused by: ", Cause->GetMessage());
        }

        return Message;
    }

private:
    TErrCode ErrorCode;
    std::string Message;
    std::shared_ptr<Error> Cause;
};
