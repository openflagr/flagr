# JSON flag source

Flagr can serve flags from a **JSON file or URL** instead of a database. The evaluation engine is identical; what changes is the authoring workflow. Flags live in a file you control, edits happen through pull requests, and a running Flagr instance becomes a read-only consumer that polls for updates. Because `json_file` and `json_http` both put the server into eval-only mode, the CRUD UI, exposure endpoint, and database are gone - only evaluation, health, and the export endpoint remain. The behavioral rules for that mode are on [Behavioral contracts](contracts.md#eval-only).

## Quick start

The smallest valid file is an empty flags array:

```json
{ "Flags": [] }
```

Point Flagr at it with two environment variables - the database driver and the connection string - and start the server:

```sh
export FLAGR_DB_DBDRIVER=json_file
export FLAGR_DB_DBCONNECTIONSTR=/path/to/flags.json
./flagr
```

If you already have a Flagr instance backed by a database, you do not need to hand-write your first file. The export endpoint returns the same JSON shape this driver reads, so you can capture your current flags and commit them to Git in one step:

```sh
curl http://localhost:18000/api/v1/export/eval_cache/json -o flags.json
git add flags.json && git commit -m "update flags"
export FLAGR_DB_DBDRIVER=json_file   # or json_http
export FLAGR_DB_DBCONNECTIONSTR=/path/to/flags.json
```

From there on, the file is the source of truth. Flagr re-fetches it on its EvalCache refresh interval, so a committed change reaches the server within that window - no restart, no deploy. The exact freshness contract and the `FLAGR_EVALCACHE_REFRESHINTERVAL` knob are on [EvalCache freshness](contracts.md#evalcache-freshness).

## Validation

A hand-edited file can have typos, a missing key, or a distribution that sums to 99 instead of 100. The server catches these at reload time - validation errors block the load, warnings log only - but you want to catch them earlier, in CI, before a bad commit reaches production. The `flagr-validate` binary runs the exact same `ValidateFlags()` the server uses, so what passes CI is what the server will accept:

```sh
go build -o flagr-validate ./cmd/flagr-validate/
./flagr-validate flags.json
```

It checks the JSON shape, required fields, key uniqueness, distribution sums (**100** when one or more distributions are present), variant references, constraint operator validity, and percent ranges. The exit code is `0` when the file is valid (warnings allowed), `1` on errors, and `2` on usage mistakes. One subtlety: `Tag.Value` is declared required by the schema but is not enforced by `ValidateFlags`, so an empty tag value will load without complaint. For programmatic use, `ValidateFlags()` is exported from the handler package.

## GitOps with GitHub

The full GitOps loop is: **author** flags in a Git repository → **review** every change in a pull request → **validate** in CI with `flagr-validate` → **serve** via `json_http` pointed at the raw file URL. Flagr polls that URL on its refresh interval, so a merged PR reaches the server without a deploy. If a change is wrong, rollback is a `git revert` - the same one-command undo you already trust for code.

### Setup

You need a fine-grained personal access token with **Contents: read** scope on the config repository, and a Flagr instance pointed at the raw content URL:

1. Create a fine-grained PAT with repo Contents read permission.
2. Point Flagr at the raw file:

   ```sh
   export FLAGR_DB_DBDRIVER=json_http
   export FLAGR_DB_DBCONNECTIONSTR="https://<PAT>@raw.githubusercontent.com/<owner>/<repo>/<ref>/flags.json"
   ```

   The PAT travels as the HTTP Basic username with an empty password. For a private repo on `main`, the connection string looks like:

   ```sh
   export FLAGR_DB_DBCONNECTIONSTR="https://github_pat_xxxx@raw.githubusercontent.com/myorg/flagr-config/main/flags.json"
   ```

### Security

Treat the token like any other secret: grant the narrowest scope that still reads the config repo, `chmod 600` any env file on shared hosts, and rotate tokens on a schedule. Because the server only needs read access, a leaked token cannot mutate your flags - it can only expose them, and rotation is a single environment variable change.

## JSON format

The file mirrors Flagr's entity model directly: a single `Flags` array at the root, each flag carrying its own segments, variants, constraints, distributions, and tags as nested objects. This is a hand-edited (or machine-generated) artifact, not a database dump you have to round-trip through an API. IDs are optional - the server assigns them on load - and distributions can reference variants by their string key instead of a numeric ID, so the file stays readable and diff-friendly even when you reorder or rename things.

The root object contains a single `Flags` array:

```json
{
  "Flags": [ ... ]
}
```

### IDs are optional

Every entity ID - flag, variant, segment, constraint, distribution, tag - is **auto-assigned** if you leave it at zero. Hand-edited files can omit IDs entirely and let the server fill them in. If you do supply IDs, they must be globally unique within their entity type, because the server uses one global counter per type to match the auto-increment behavior of a real database. That counter also means IDs stay stable if you ever migrate the file back to a SQL backend.

Distributions can name their target variant by `VariantKey` instead of `VariantID`. The loader resolves the key to the variant's assigned ID on read, so you never have to manage numeric cross-references by hand.

> **Note:** `SegmentDefaultRank` (`999`) is applied by the **CRUD API** when
> creating segments, **not** by the JSON loader. If you omit `Rank` in a JSON
> flag, it stays `0` - set it explicitly when segment order matters.

### Flag

A flag is the top-level unit. It carries a unique key for evaluation requests, a human description, an enabled toggle, and the nested arrays that define its audience and outcomes. Notes are freeform markdown (the UI renders KaTeX). `DataRecordsEnabled` and `EntityType` control whether and how evaluation results are logged to the metrics pipeline.

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
| `DataRecordsEnabled` | bool | no | Log evaluation data to the metrics pipeline |
| `EntityType` | string | no | Override entity type in evaluation logs |

### Variant

Variants are the possible outcomes of an evaluation. Each has a unique key within the flag and an optional attachment - arbitrary JSON the client receives alongside the variant key, useful for feature configuration like colors, layouts, or feature flags passed downstream.

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

Segments divide a flag's audience into ranked groups. `Rank` sets evaluation priority - lower ranks evaluate first, so the first segment whose **constraints match** wins and evaluation stops (rollout miss does not fall through). `RolloutPercent` gates what fraction of that matched segment's hashed range gets a variant. Details: [contracts: segment evaluation](contracts.md#segment-evaluation).

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
| `Rank` | uint | no | Evaluation priority (lower = higher priority). **Defaults to `0` in JSON** (the `999` default only applies to the CRUD API) |
| `RolloutPercent` | uint | no | Percentage of users matching this segment (`0-100`) |
| `Constraints` | array | no | Conditions that must match |
| `Distributions` | array | no | How to route matched users across variants |

### Constraint

A constraint is a single condition on one entity property. `Value` is a JSON-encoded string - note the nested quotes for string literals - so the same field can carry a number, a string, or a list without a schema change.

```json
{
  "Property": "country",
  "Operator": "EQ",
  "Value": "\"US\""
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Property` | string | yes | Entity property to evaluate (e.g. `"country"`, `"age"`) |
| `Operator` | string | yes | Comparison operator (see below) |
| `Value` | string | yes | Value to compare against (JSON-encoded) |

**Operators** (12 supported):

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

A distribution routes a share of a segment's traffic to one variant. Use `VariantKey` to name the target by its string key, or `VariantID` if you prefer the numeric form - exactly one is required. The `Percent` values across all distributions in a segment must sum to **100** when at least one distribution exists; a segment with zero distributions yields a warning instead.

```json
{
  "VariantKey": "control",
  "Percent": 50
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `VariantKey` | string | yes* | Target variant key |
| `VariantID` | uint | yes* | Target variant ID (alternative to `VariantKey`) |
| `Percent` | uint | yes | Percentage of segment traffic (`0-100`). **Must sum to 100 across all distributions in a segment** (enforced when ≥1 distribution exists; zero distributions yields a warning) |

*Either `VariantKey` or `VariantID` is required.

### Tag

Tags are freeform labels for grouping and searching flags. A flag can carry any number of them, and the evaluation API can filter by tag.

```json
{
  "Value": "frontend"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Value` | string | yes¹ | Tag value |

¹ `Value` is required by the schema, but `ValidateFlags` does not enforce it - 
an empty value loads without error.

## Complete example

Two flags in one file, no explicit IDs anywhere - the server assigns them on load. The first rolls a new dashboard out to 50% of users via a single segment with an even 50/50 split. The second gates maintenance mode, sending beta users to the `on` variant first and everyone else to `off`.

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