# JSON Flag Source

Flagr can load flags from a JSON file instead of a database. This is the foundation for GitOps workflows — manage flags as code, validate before deploy, and let Flagr serve them.

## Quick start

**From scratch** — create a file and point Flagr at it:

```json
{ "Flags": [] }
```

```sh
export FLAGR_DB_DBDRIVER=json_file
export FLAGR_DB_DBCONNECTIONSTR=/path/to/flags.json
./flagr
```

**From an existing instance** — export, commit, deploy:

```sh
# Export from a running Flagr
curl http://localhost:18000/api/v1/export/eval_cache/json -o flags.json

# Edit, commit, push
git add flags.json && git commit -m "update flags"

# Deploy via local file or HTTP
export FLAGR_DB_DBDRIVER=json_file       # or json_http
export FLAGR_DB_DBCONNECTIONSTR=/path/to/flags.json
```

Flagr reloads flags automatically on the cache refresh interval (default: 3 seconds).

## Validation

Validate your flag file before deploying:

```sh
go build -o flagr-validate ./cmd/flagr-validate/
./flagr-validate flags.json
```

The validator checks: valid JSON, required fields, key uniqueness, distribution sums (must be 100), variant references, constraint expressions, and percentage ranges. It reports errors (must fix) and warnings (should fix) separately.

You can also use `ValidateFlags()` from `pkg/handler` programmatically.
## GitOps with GitHub

Host your `flags.json` in a Git repository and point Flagr at the raw file. This gives you full GitOps: PR review, audit trail, rollback via `git revert`, and CI validation before deploy.

### Setup

1. **Create a GitHub Personal Access Token** (fine-grained, repo-scoped):
   - Go to **Settings → Developer settings → Personal access tokens → Fine-grained tokens**
   - Scope to the repository containing your flags file
   - Grant **Read access to Contents**

2. **Point Flagr at the raw file** using `json_http` with the token embedded in the URL:

   ```sh
   export FLAGR_DB_DBDRIVER=json_http
   export FLAGR_DB_DBCONNECTIONSTR="https://<PAT>@raw.githubusercontent.com/<owner>/<repo>/<ref>/flags.json"
   ```

   The token is used as HTTP Basic Auth username (the password is empty), which Go's `net/http` handles transparently. GitHub accepts this for raw content access.

   **Example** — load from a private repo's `main` branch:

   ```sh
   export FLAGR_DB_DBCONNECTIONSTR="https://github_pat_xxxx@raw.githubusercontent.com/myorg/flagr-config/main/flags.json"
   ```

### Security notes

- Use a **fine-grained token** with the narrowest scope possible (single repo, read-only Contents).
- The token is visible in the environment and process listing. On shared hosts, restrict access to the env file (e.g., `chmod 600`).
- Consider a dedicated machine account for the token rather than a personal account.
- Rotate tokens on a schedule; GitHub fine-grained tokens support expiration.

### CI validation

Validate your flag file in CI before merges land on the branch Flagr watches:

```sh
go build -o flagr-validate ./cmd/flagr-validate/
./flagr-validate flags.json
```

Failing validation blocks the PR — broken flag config never reaches your running instances.

## JSON format

The root object contains a single `Flags` array:

```json
{
  "Flags": [ ... ]
}
```

### IDs are optional

All entity IDs (flags, variants, segments, constraints, distributions, tags) are **auto-assigned** if omitted. This means hand-edited files can skip IDs entirely. If you do provide them, they must be globally unique per entity type.

Distributions can reference variants by `VariantKey` instead of `VariantID` — the system resolves the link automatically.

### Flag

```json
{
  "Key": "my-feature",
  "Description": "Controls the new dashboard rollout",
  "Enabled": true,
  "Segments": [ ... ],
  "Variants": [ ... ],
  "Tags": [ ... ],
  "Notes": "Optional markdown notes",
  "DataRecordsEnabled": true,
  "EntityType": "user"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Key` | string | yes | Unique key for evaluation requests |
| `Description` | string | no | Human-readable description |
| `Enabled` | bool | no | Whether the flag is active |
| `Segments` | array | no | Audience segments |
| `Variants` | array | no | Possible evaluation outcomes |
| `Tags` | array | no | Searchable tags |
| `Notes` | string | no | Markdown notes (supports KaTeX in the UI) |
| `DataRecordsEnabled` | bool | no | Log evaluation data to metrics pipeline |
| `EntityType` | string | no | Override entity type in evaluation logs |

### Variant

```json
{
  "Key": "control",
  "Attachment": { "color": "blue" }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Key` | string | yes | Unique key within the flag |
| `Attachment` | object | no | Arbitrary JSON configuration for this variant |

### Segment

```json
{
  "Description": "All US users",
  "Rank": 0,
  "RolloutPercent": 100,
  "Constraints": [ ... ],
  "Distributions": [ ... ]
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Description` | string | no | Human-readable description |
| `Rank` | uint | no | Evaluation priority (lower = higher priority). Default: 999 |
| `RolloutPercent` | uint | no | Percentage of users matching this segment (0-100) |
| `Constraints` | array | no | Conditions that must match |
| `Distributions` | array | no | How to route matched users across variants |

### Constraint

```json
{
  "Property": "country",
  "Operator": "EQ",
  "Value": "\"US\""
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Property` | string | yes | Entity property to evaluate (e.g., `"country"`, `"age"`) |
| `Operator` | string | yes | Comparison operator (see below) |
| `Value` | string | yes | Value to compare against |

**Operators:**

| Operator | Description | Example Value |
|----------|-------------|---------------|
| `EQ` | Equal | `"\"US\""` |
| `NEQ` | Not equal | `"\"US\""` |
| `LT` | Less than | `"25"` |
| `LTE` | Less than or equal | `"25"` |
| `GT` | Greater than | `"18"` |
| `GTE` | Greater than or equal | `"18"` |
| `EREG` | Regex match | `"\"^US.*\""` |
| `NEREG` | Regex not match | `"\"^US.*\""` |
| `IN` | Value in list | `"[\"US\", \"CA\", \"UK\"]"` |
| `NOTIN` | Value not in list | `"[\"US\", \"CA\", \"UK\"]"` |
| `CONTAINS` | String contains | `"\"california\""` |
| `NOTCONTAINS` | String not contains | `"\"california\""` |

### Distribution

```json
{
  "VariantKey": "control",
  "Percent": 50
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `VariantKey` | string | yes* | Target variant key |
| `VariantID` | uint | yes* | Target variant ID (alternative to VariantKey) |
| `Percent` | uint | yes | Percentage of segment traffic (0-100). **Must sum to 100 across all distributions in a segment.** |

*Either `VariantKey` or `VariantID` is required.

### Tag

```json
{
  "Value": "frontend"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Value` | string | yes | Tag value |

## Complete example
Two flags, no explicit IDs — the system auto-assigns them on load.
```json
{
  "Flags": [
    {
      "Key": "new-dashboard",
      "Description": "Controls the new dashboard rollout",
      "Enabled": true,
      "EntityType": "user",
      "DataRecordsEnabled": false,
      "Notes": "Rolling out new dashboard to 50% of users",
      "Tags": [
        { "Value": "frontend" },
        { "Value": "experiment" }
      ],
      "Variants": [
        {
          "Key": "control",
          "Attachment": { "color": "blue", "layout": "classic" }
        },
        {
          "Key": "treatment",
          "Attachment": { "color": "green", "layout": "modern" }
        }
      ],
      "Segments": [
        {
          "Description": "All users",
          "Rank": 0,
          "RolloutPercent": 100,
          "Constraints": [],
          "Distributions": [
            { "VariantKey": "control", "Percent": 50 },
            { "VariantKey": "treatment", "Percent": 50 }
          ]
        }
      ]
    },
    {
      "Key": "maintenance-mode",
      "Description": "Enables maintenance mode for the API",
      "Enabled": false,
      "EntityType": "request",
      "DataRecordsEnabled": true,
      "Tags": [
        { "Value": "ops" }
      ],
      "Variants": [
        { "Key": "off", "Attachment": {} },
        {
          "Key": "on",
          "Attachment": { "message": "System maintenance in progress", "retryAfter": 300 }
        }
      ],
      "Segments": [
        {
          "Description": "Beta users get maintenance mode early",
          "Rank": 0,
          "RolloutPercent": 100,
          "Constraints": [
            { "Property": "tier", "Operator": "EQ", "Value": "\"beta\"" }
          ],
          "Distributions": [
            { "VariantKey": "on", "Percent": 100 }
          ]
        },
        {
          "Description": "All other users",
          "Rank": 1,
          "RolloutPercent": 100,
          "Constraints": [],
          "Distributions": [
            { "VariantKey": "off", "Percent": 100 }
          ]
        }
      ]
    }
  ]
}
```
