# chaosplane-platform

Proprietary SaaS platform for Chaosplane — chaos engineering experiments as a service.

## Structure

```
apps/
  api/       Go API server (Gin + Wire DI + slog)
```

## Development

```bash
# Build the API server
make build-api

# Run tests
make test-api

# Lint
make lint-api
```

## License

Proprietary. See [LICENSE](LICENSE).
