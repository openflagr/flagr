---
title: AI & MCP Integration
---

# AI & MCP Integration

Flagr ships with a built-in **MCP (Model Context Protocol) server** that lets AI coding agents manage feature flags directly. Instead of opening the UI or calling curl, you ask your agent to create a flag, add a segment, or run an evaluation — and it does.

The MCP server runs inside the same Flagr binary and serves requests over HTTP at `/mcp`. No separate process, no extra ports. The tools call Flagr's internal domain layer directly.

## Quick Start

**1. Start Flagr with MCP enabled**

```bash
FLAGR_MCP_ENABLED=true ./flagr
```

The MCP endpoint is available at `http://localhost:18000/mcp`. The HTTP API continues to work as normal on the same port.

**2. Point your AI tool at it**

Each IDE / agent has its own config shape. Pick one from the [Integration Guides](#integration-guides) below and paste the snippet.

**3. Ask your agent**

> "Create a flag called `new-checkout` with a `control` and `treatment` variant, then add a 50/50 distribution."

The agent calls the MCP tools, Flagr does the work, and you get the result as structured JSON.

## Tool Reference

All 27 tools are listed below, grouped by domain. Every tool returns JSON (or an error message with `isError: true`).

### Flag Management

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `create_flag` | Create a new feature flag | — | `key`, `description`, `enabled`, `data_records_enabled`, `entity_type`, `notes` |
| `get_flag` | Get a flag by ID with full detail (segments, variants, tags) | `flag_id` | — |
| `list_flags` | List flags with optional filters | — | `enabled`, `description_like`, `tags`, `preload`, `limit`, `offset` |
| `update_flag` | Update a flag's properties | `flag_id` | `description`, `key`, `enabled`, `data_records_enabled`, `entity_type`, `notes` |
| `set_flag_enabled` | Enable or disable a flag | `flag_id`, `enabled` | — |
| `delete_flag` | Soft-delete a flag | `flag_id` | — |
| `duplicate_flag` | Duplicate a flag with all segments, variants, tags, constraints, and distributions | `flag_id` | — |

### Segment Management

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `create_segment` | Create a segment on a flag | `flag_id`, `rollout_percent` | `description` |
| `list_segments` | List segments for a flag, ordered by rank | `flag_id` | — |
| `update_segment` | Update a segment's rollout percent and description | `flag_id`, `segment_id` | `rollout_percent`, `description` |
| `delete_segment` | Delete a segment from a flag | `flag_id`, `segment_id` | — |

### Variant Management

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `create_variant` | Create a variant on a flag | `flag_id`, `key` | `attachment` |
| `list_variants` | List all variants for a flag | `flag_id` | — |
| `update_variant` | Update a variant's key and attachment | `flag_id`, `variant_id` | `key`, `attachment` |
| `delete_variant` | Delete a variant from a flag | `flag_id`, `variant_id` | — |

### Constraint & Distribution

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `create_constraint` | Create a constraint on a segment | `flag_id`, `segment_id`, `property`, `operator`, `value` | — |
| `list_constraints` | List constraints for a segment | `flag_id`, `segment_id` | — |
| `update_constraint` | Update a constraint's property, operator, or value | `flag_id`, `constraint_id` | `property`, `operator`, `value` |
| `delete_constraint` | Delete a constraint from a segment | `flag_id`, `constraint_id` | — |
| `list_distributions` | List distributions for a segment | `flag_id`, `segment_id` | — |
| `set_distributions` | Overwrite all distributions for a segment (percents must sum to 100) | `flag_id`, `segment_id`, `distributions` | — |

### Tags

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `list_tags` | List tags on a flag | `flag_id` | — |
| `add_tag` | Add a tag to a flag (max 64 chars, unique per flag) | `flag_id`, `value` | — |
| `delete_tag` | Remove a tag from a flag | `flag_id`, `tag_id` | — |
| `list_all_tags` | List all tags across all flags | — | `value_like`, `limit`, `offset` |

### Evaluation & Health

| Tool | Description | Required Params | Optional Params |
|------|-------------|-----------------|-----------------|
| `evaluate_flag` | Evaluate a flag for a given entity. Returns the matching variant. | `entity_id` | `flag_id`, `flag_key`, `entity_type`, `entity_context` |
| `health` | Check Flagr server health | — | — |

### Constraint Operators

The `operator` field on constraints accepts:

| Operator | Meaning | Value Format |
|----------|---------|--------------|
| `EQ` | Equals | `"foo"` |
| `NEQ` | Not equals | `"foo"` |
| `LT` | Less than | `"100"` |
| `LTE` | Less than or equal | `"100"` |
| `GT` | Greater than | `"100"` |
| `GTE` | Greater than or equal | `"100"` |
| `EREG` | Regex match | `"(us\|ca)"` |
| `NEREG` | Regex does not match | `"^internal"` |
| `IN` | Value in set | `"a","b","c"` |
| `NOTIN` | Value not in set | `"x","y"` |
| `CONTAINS` | String contains | `"premium"` |
| `NOTCONTAINS` | String does not contain | `"beta"` |

Strings must be double-quoted. Arrays use comma-separated quoted values. Regexes are EREG-style.

## Integration Guides

### OpenCode

Add the MCP server to your `opencode.json` (project root) or `~/.config/opencode/opencode.json` (global):

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "flagr": {
      "type": "remote",
      "url": "http://localhost:18000/mcp",
      "enabled": true
    }
  }
}
```

The agent will see all 27 Flagr tools in its tool list. You can also use the CLI config flow — run `/help` inside opencode for details.

### Cursor

Create `.cursor/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "flagr": {
      "url": "http://localhost:18000/mcp",
      "transport": "http"
    }
  }
}
```

Cursor auto-detects `.cursor/mcp.json`. The tools appear in the MCP panel and are available to the agent during chat.

### Zed

Add to `settings.json` (open via `Cmd+,`):

```json
{
  "context_servers": {
    "flagr": {
      "url": "http://localhost:18000/mcp"
    }
  }
}
```

### Claude Code

**CLI (recommended):**

```bash
claude mcp add --transport http flagr http://localhost:18000/mcp
```

**Config file (`.claude/settings.json`):**

```json
{
  "mcpServers": {
    "flagr": {
      "url": "http://localhost:18000/mcp"
    }
  }
}
```

### Windsurf

Create `.windsurfrules` or use the MCP configuration in your Windsurf settings:

```json
{
  "mcpServers": {
    "flagr": {
      "url": "http://localhost:18000/mcp"
    }
  }
}
```

### Cline (VS Code Extension)

Open Cline settings (`Cmd+Shift+P` → "Cline: Open MCP Settings") and add:

```json
{
  "mcpServers": {
    "flagr": {
      "command": "/path/to/flagr",
      "args": [],
      "env": {
        "FLAGR_MCP_ENABLED": "true"
      }
    }
  }
}
```

### Continue.dev

Add to `config.yaml` (usually `~/.continue/config.yaml`):

```yaml
mcpServers:
  - name: flagr
    url: http://localhost:18000/mcp
```

## Usage Examples

### Creating a flag with variants and segments

**What you'd ask your agent:**

> "Create a flag for the new checkout flow. Add a `control` variant and a `treatment` variant. Create a segment for US users at 50% rollout."

**What the agent does (tool calls):**

1. `create_flag` → `{ "key": "new-checkout", "description": "New checkout flow" }`
2. `create_variant` → `{ "flag_id": 1, "key": "control" }`
3. `create_variant` → `{ "flag_id": 1, "key": "treatment", "attachment": { "ui_version": "v2" } }`
4. `create_segment` → `{ "flag_id": 1, "rollout_percent": 50, "description": "US users" }`
5. `create_constraint` → `{ "flag_id": 1, "segment_id": 1, "property": "country", "operator": "EQ", "value": "\"US\"" }`
6. `set_distributions` → `{ "flag_id": 1, "segment_id": 1, "distributions": [{ "variant_id": 1, "variant_key": "control", "percent": 50 }, { "variant_id": 2, "variant_key": "treatment", "percent": 50 }] }`

### Setting up a gradual rollout with constraints

**What you'd ask your agent:**

> "I want to gradually roll out the new recommendation engine. Start with internal employees at 100%, then beta users at 25%, and eventually everyone."

**What the agent does:**

1. Creates the flag and variants
2. Creates segment "Internal" (rank 1, 100% rollout) with a constraint: `email ENDSWITH "@yourcompany.com"` — or `property: "role", operator: "EQ", value: "\"internal\""`
3. Creates segment "Beta users" (rank 2, 25% rollout) with constraint: `beta == true`
4. Creates segment "Everyone else" (rank 3, 100% rollout) with no constraints (catch-all)

### Evaluating flags from code

The `evaluate_flag` tool lets you test evaluations without writing code:

```json
{
  "flag_key": "new-checkout",
  "entity_id": "user-123",
  "entity_type": "user",
  "entity_context": {
    "country": "US",
    "plan": "pro"
  }
}
```

Returns:

```json
{
  "flag_id": 1,
  "flag_key": "new-checkout",
  "variant_id": 2,
  "variant_key": "treatment",
  "variant_attachment": { "ui_version": "v2" },
  "segment_id": 1,
  "debug_msg": "matched segment: US users"
}
```

### Managing tags for batch evaluation

Tags let you group flags for batch evaluation. Add them as you create flags:

1. `add_tag` → `{ "flag_id": 1, "value": "checkout" }`
2. `add_tag` → `{ "flag_id": 1, "value": "experiment" }`
3. `list_flags` → `{ "tags": "checkout,experiment" }` — returns all flags matching those tags

### Duplicating a flag across environments

Use `duplicate_flag` to clone a flag with all its configuration (segments, variants, constraints, distributions, tags) into a new flag. The duplicate is always created disabled so you can review before enabling:

```json
{ "flag_id": 1 }
```

Then update the duplicate's key and enable it when ready:

```json
{ "flag_id": 5, "key": "new-checkout-staging", "enabled": true }
```

## Architecture

```
┌─────────────────────────────────────────┐
│              Flagr Binary                │
│                                         │
│         ┌────────────────┐              │
│         │   HTTP Server   │              │
│         │   :18000        │              │
│         └───┬────────┬───┘              │
│             │        │                   │
│    /api/v1/*    /mcp (MCP)               │
│             │        │                   │
│         ┌───▼────────▼───┐              │
│         │  Handler Layer   │              │
│         │  (EvalCache)     │              │
│         └────────┬────────┘              │
│                  │                       │
│         ┌────────▼────────┐              │
│         │  Entity / ORM    │              │
│         │  (GORM)          │              │
│         └────────┬────────┘              │
│                  │                       │
│         ┌────────▼────────┐              │
│         │  Database        │              │
│         │  (SQLite/MySQL/  │              │
│         │   Postgres)      │              │
│         └─────────────────┘              │
└─────────────────────────────────────────┘
```

Key points:

- **Built-in, not a sidecar.** The MCP server lives in the same process as the HTTP server. One binary, one deployment, one port.
- **Streamable HTTP transport.** The MCP server is mounted at `/mcp` on the same HTTP port as the API. Uses the MCP Streamable HTTP protocol (JSON-RPC over HTTP).
- **Direct domain calls.** MCP tools call the same handler/entity layer as the HTTP API. No HTTP loopback, no latency overhead.
- **Evaluation cache.** The `evaluate_flag` tool uses the same in-memory EvalCache as the HTTP evaluation endpoint. The cache is started automatically when Flagr boots.
- **Audit trail.** All MCP-created entities are stamped with `created_by: "mcp"` and `updated_by: "mcp"`.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `FLAGR_MCP_ENABLED` | `false` | Set to `true` to enable the MCP server |
| `FLAGR_DB_DBDRIVER` | `sqlite3` | Database driver (sqlite3, mysql, postgres) |
| `FLAGR_DB_DBCONNECTIONSTR` | `flagr.sqlite` | Database connection string |
| `HOST` | `localhost` | HTTP API host |
| `PORT` | `18000` | HTTP API port |

The MCP server uses no additional ports or environment variables beyond `FLAGR_MCP_ENABLED`. All other configuration (database, auth, logging) is shared with the HTTP API.

## Troubleshooting

### The MCP server doesn't start

- Verify `FLAGR_MCP_ENABLED=true` is set. The server is off by default.
- Check Flagr is running: `curl http://localhost:18000/api/v1/health`
- Check MCP endpoint: `curl -X POST http://localhost:18000/mcp -H 'Content-Type: application/json' -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'`

### No tools appear in my AI tool

- Restart your IDE / agent after adding the MCP config.
- Check the MCP server logs in your tool's output panel.
- Verify the MCP endpoint is reachable: `curl http://localhost:18000/mcp`

### Evaluation returns empty variantKey

This is expected behavior, not an error. It means:
- The flag is disabled (`enabled: false`), or
- The entity matched a segment but rollout excluded them, or
- No segment matched at all.

Use `evaluate_flag` with full `entity_context` to debug. The response includes `debug_msg` with the evaluation trace.

### "flag not found" errors

Ensure you're using the correct `flag_id` (integer) returned by `create_flag` or `list_flags`. Flag IDs are auto-incrementing integers, not keys.

### Permission errors in production

The MCP server has the same access as the HTTP API. If you have auth enabled (JWT, Basic, etc.), those settings apply to the HTTP API only — the MCP server bypasses HTTP auth because it calls the domain layer directly. Consider this when enabling MCP in production environments.
