# AGENTS.md

Flagr is a Go feature flag and A/B testing service with REST APIs.

## Commands

```bash
make test          # Run tests
make build         # Build server
make start         # Run backend + frontend dev server
make swagger       # Regenerate API code
```

## Key Code

- Handlers: `pkg/handler/crud.go`, `pkg/handler/eval.go`
- Entities: `pkg/entity/`

## Notes

- **Don't edit `swagger_gen/`** - regenerate with `make swagger`
- Dev uses SQLite
- See [deepwiki.com/openflagr/flagr](https://deepwiki.com/openflagr/flagr) and `docs/`
