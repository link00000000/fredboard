class HTTPError(RuntimeError):
    pass

class UnauthorizedError(HTTPError):
    pass

class RateLimitError(HTTPError):
    pass

