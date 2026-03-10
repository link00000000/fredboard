#pragma once

#include <expected>
#include <memory>
#include <sstream>
#include <string>
#include <type_traits>
#include <utility>
#include <variant>

namespace Core
{
    struct ICause
    {
        virtual ~ICause() = default;
        virtual std::string GetFullTrace() const = 0;
    };

    template <typename TErrCode, const char* const* TErrNames>
    struct ErrorCause;

    template <typename TErrCode, const char* const* TErrNames>
    requires std::is_enum_v<TErrCode>
    struct Error
    {
        Error(TErrCode InErrCode)
            : ErrorCode(InErrCode)
        {
        }

        Error(TErrCode InErrCode, std::string InMessage)
            : ErrorCode(InErrCode)
            , Message(std::move(InMessage))
        {
        }

        template <typename TCauseCode, const char* const* TCauseNames>
        Error(TErrCode InErrCode, const Error<TCauseCode, TCauseNames>& InCause)
            : ErrorCode(InErrCode)
            , Cause(std::make_shared<ErrorCause<TCauseCode, TCauseNames>>(InCause))
        {
        }

        template <typename TCauseCode, const char* const* TCauseNames>
        Error(TErrCode InErrCode, std::string InMessage, const Error<TCauseCode, TCauseNames>& InCause)
            : ErrorCode(InErrCode)
            , Message(std::move(InMessage))
            , Cause(std::make_shared<ErrorCause<TCauseCode, TCauseNames>>(InCause))
        {
        }

        bool Is(TErrCode InErrCode) const { return ErrorCode == InErrCode; }
        std::string_view GetName()    const { return TErrNames[std::to_underlying(ErrorCode)]; }
        std::string_view GetMessage() const { return Message; }

        std::string GetFullTrace() const
        {
            if (Cause)
            {
                return std::format("({}) {}\ncaused by: {}", GetName(), Message, Cause->GetFullTrace());
            }

            return std::format("({}) {}", GetName(), Message);
        }

    private:
        TErrCode ErrorCode;
        std::string Message;
        std::shared_ptr<ICause> Cause; // type-erased cause, only for tracing
    };

    template <typename TErrCode, const char* const* TErrNames>
    struct ErrorCause : ICause
    {
        explicit ErrorCause(Error<TErrCode, TErrNames> InError)
            : Inner(std::move(InError))
        {
        }

        std::string GetFullTrace() const override { return Inner.GetFullTrace(); }

    private:
        Error<TErrCode, TErrNames> Inner;
    };

    template <typename TValue, typename... TErrors>
    struct Result
    {
        using VariantType = std::variant<TErrors...>;
        using ExpectedType = std::expected<TValue, VariantType>;

        Result(TValue InValue)
            : Inner(std::move(InValue))
        {
        }

        template <typename TError>
        requires (std::is_same_v<TError, TErrors> || ...)
        Result(TError InError)
            : Inner(std::unexpected(VariantType(std::move(InError))))
        {
        }

        template <typename TError>
        requires (std::is_same_v<TError, TErrors> || ...)
        Result(std::unexpected<TError> InError)
            : Inner(std::unexpected(VariantType(std::move(InError.error()))))
        {
        }

        bool HasValue() const { return Inner.has_value(); }
        explicit operator bool() const { return Inner.has_value(); }

        TValue&       Value()       { return Inner.value(); }
        const TValue& Value() const { return Inner.value(); }

        TValue& operator*()              { return *Inner; }
        const TValue& operator*()  const { return *Inner; }
        TValue* operator->()             { return Inner.operator->(); }
        const TValue* operator->() const { return Inner.operator->(); }

        const VariantType& Error() const { return Inner.error(); }

        template <typename... THandlers>
        auto MatchError(THandlers&&... Handlers) const
        {
            static_assert((std::is_invocable_v<THandlers, const TErrors&> && ...), "Not all error types are handled");
            return std::visit(ErrorVisitor<std::decay_t<THandlers>...>{ std::forward<THandlers>(Handlers)... }, Inner.error());
        }

        template <typename THandler>
        auto AnyError(THandler&& Handler) const
        {
            return std::visit([&Handler](const auto& Err) { return Handler(Err); }, Inner.error());
        }

        template <typename TOuterError, typename... TArgs>
        auto WrapError(TArgs&&... Args) const
        {
            return std::visit([&](const auto& Err) { return TOuterError(std::forward<TArgs>(Args)..., Err); }, Inner.error());
        }

    private:
        template <typename... Ts>
        struct ErrorVisitor : Ts...
        {
            using Ts::operator()...;

            template <typename TUnexpected>
            void operator()(const TUnexpected&) const
            {
                static_assert(
                    !sizeof(TUnexpected),
                    "MatchError: handler for type that is not in the error variant -- see TUnexpected in the call stack"
                );
            }
        };

        ExpectedType Inner;
    };

    // TODO: Delete?
    template <typename TErrCode, const char* const* TErrNames>
    requires std::is_enum_v<TErrCode>
    struct ErrorOld
    {
        ErrorOld(TErrCode InErrCode, std::string InErrorMessage)
            : ErrorCode(InErrCode), Message(std::move(InErrorMessage)), Cause(nullptr)
        {
        }

        ErrorOld(TErrCode InErrCode, std::string InErrorMessage, const ErrorOld& InCause)
            : ErrorCode(InErrCode), Message(std::move(InErrorMessage)), Cause(std::make_shared<ErrorOld>(std::move(InCause)))
        {
        }

        ErrorOld(TErrCode InErrCode, const ErrorOld& InCause)
            : ErrorCode(InErrCode), Cause(std::make_shared<ErrorOld>(std::move(InCause)))
        {
        }

        bool Is(TErrCode InErrCode) const
        {
            if (ErrorCode == InErrCode)
            {
                return true;
            }

            if (Cause)
            {
                return Cause.Is(InErrCode);
            }

            return false;
        }

        std::string GetName() const
        {
            return TErrNames[std::to_underlying(ErrorCode)];
        }

        const std::string& GetMessage() const
        {
            if (Cause)
            {
                return std::format("({}) {}\ncaused by: {}", GetName(), Message, Cause->GetMessage());
            }

            return Message;
        }

    private:
        TErrCode ErrorCode;
        std::string Message;
        std::shared_ptr<ErrorOld> Cause;
    };
}

#define PRIVATE_DECLARE_ERROR_STR(X_) #X_,

#define PRIVATE_DECLARE_ERROR_FE_1(Func_, Arg1_)                Func_(Arg1_)
#define PRIVATE_DECLARE_ERROR_FE_2(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_1(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_3(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_2(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_4(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_3(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_5(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_4(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_6(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_5(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_7(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_6(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_8(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_7(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_9(Func_, Arg1_,...)            Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_8(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_10(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_9(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_11(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_10(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_12(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_11(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_13(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_12(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_14(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_13(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_15(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_14(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_16(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_15(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_17(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_16(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_18(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_17(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_19(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_18(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_20(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_19(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_21(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_20(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_22(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_21(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_23(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_22(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_24(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_23(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_25(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_24(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_26(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_25(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_27(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_26(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_28(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_27(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_29(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_28(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_30(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_29(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_31(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_30(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_32(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_31(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_33(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_32(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_34(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_33(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_35(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_34(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_36(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_35(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_37(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_36(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_38(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_37(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_39(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_38(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_40(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_39(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_41(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_40(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_42(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_41(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_43(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_42(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_44(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_43(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_45(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_44(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_46(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_45(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_47(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_46(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_48(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_47(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_49(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_48(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_50(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_49(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_51(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_50(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_52(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_51(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_53(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_52(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_54(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_53(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_55(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_54(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_56(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_55(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_57(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_56(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_58(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_57(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_59(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_58(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_60(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_59(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_61(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_60(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_62(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_61(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_63(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_62(Func_, __VA_ARGS__)
#define PRIVATE_DECLARE_ERROR_FE_64(Func_, Arg1_,...)           Func_(Arg1_) PRIVATE_DECLARE_ERROR_FE_63(Func_, __VA_ARGS__)

#define PRIVATE_DECLARE_ERROR_FE_SELECT(                                        \
     _1, _2, _3, _4, _5, _6, _7, _8, _9,_10,_11,_12,_13,_14,_15,_16,            \
    _17,_18,_19,_20,_21,_22,_23,_24,_25,_26,_27,_28,_29,_30,_31,_32,            \
    _33,_34,_35,_36,_37,_38,_39,_40,_41,_42,_43,_44,_45,_46,_47,_48,            \
    _49,_50,_51,_52,_53,_54,_55,_56,_57,_58,_59,_60,_61,_62,_63,_64, N, ...)    \
    PRIVATE_DECLARE_ERROR_FE_##N

#define PRIVATE_DECLARE_ERROR_FOR_EACH(Func_, ...)                              \
    PRIVATE_DECLARE_ERROR_FE_SELECT(__VA_ARGS__,                                \
        64,63,62,61,60,59,58,57,56,55,54,53,52,51,50,49,                        \
        48,47,46,45,44,43,42,41,40,39,38,37,36,35,34,33,                        \
        32,31,30,29,28,27,26,25,24,23,22,21,20,19,18,17,                        \
        16,15,14,13,12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1)(Func_, __VA_ARGS__)

#define DECLARE_ERROR(Name_, ...)                                               \
    enum class Name_##ErrorCode { __VA_ARGS__ };                                \
    inline constexpr const char* Name_##Names[] = {                             \
        PRIVATE_DECLARE_ERROR_FOR_EACH(PRIVATE_DECLARE_ERROR_STR, __VA_ARGS__)  \
    };                                                                          \
    using Name_##Error = Core::Error<Name_##ErrorCode, Name_##Names>;
