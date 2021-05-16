class HTTPError(RuntimeError):
    pass

class UnauthorizedError(HTTPError):
    pass

class RateLimitError(HTTPError):
    pass

class ConfigError(ValueError):
    pass

class GeneratedConfigError(ConfigError):
    pass

class MalformedConfigError(ConfigError):
    pass
