# log module

Handles application logging. Supports config-driven debug/info/error output, and can be extended for file or cloud logging.

## Key Types
- `InitLogger(debug bool)`: Initializes logger
- `log.Info`, `log.Error`, etc.: Logging functions

## Example
```go
log.InitLogger(true)
log.Info("message: %v", val)
```

## Extension
- Add file logging, structured logging, or cloud log sinks

