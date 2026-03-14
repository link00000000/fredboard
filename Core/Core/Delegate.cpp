#include "Delegate.h"

namespace Core
{
    DelegateHandle DelegateHandle::Invalid = DelegateHandle(0);

    bool DelegateHandle::operator==(const DelegateHandle& Other) const
    {
        return Value == Other.Value;
    }

    DelegateHandle::DelegateHandle(const int64_t InValue)
        : Value(InValue)
    {
    }

    DelegateHandle DelegateHandleGenerator::GenerateNextHandle()
    {
        return DelegateHandle(NextValue++);
    }
}
