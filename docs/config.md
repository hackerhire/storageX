# config module

Handles loading and parsing of application configuration (JSON). Supports config sections for logging, cloud providers, metadata, and more.

## Key Types
- `AppConfig`: Main config struct
- `LoadConfig(path string)`: Loads config from file

## Example
```go
cfg := config.LoadConfig("config/config.json")
fmt.Println(cfg.ChunkSize)
```

## Extension
- Add support for environment variable overrides
- Add config validation
