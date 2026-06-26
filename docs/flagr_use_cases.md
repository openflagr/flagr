# Flagr Use Cases

Feature flags, A/B tests, and dynamic configuration look like different
problems — a kill switch, a marketing experiment, a runtime knob. But at the
moment of decision, they all ask the same question: *given this entity, which
experience should it get?* That is why Flagr consolidates them into one
primitive — the **flag**. The difference between a feature flag and an
experiment is not structural, it is operational: a feature flag cares about
*who* is on, an experiment cares about *what they got* and *whether it
mattered*. The instrumentation is identical; the analysis differs. This page
shows the three patterns side by side so you can see how the same evaluation
call serves all three.

> **Note:** The pseudocode below uses the actual API response field names —
> `variantID`, `variantKey`, and `variantAttachment` (camelCase), as returned
> by `POST /evaluation`.

## Feature flagging

A feature flag is the simplest decision in software: should this code path
run, or not? The most familiar shape is the **kill switch** — a globally
visible "off" you can throw the instant a deploy goes wrong, without
redeploying or paging an engineer to roll back. But the same mechanism
generalizes to **targeted rollouts**: turn a feature on for internal users
first, then a single region, then everyone. The key property is that *deploy*
and *release* become independent — you can ship code to production hours
before any user sees it.

Given an entity (user, request, or cookie), Flagr evaluates it against the
flag:

```js
evaluation_result = flagr.postEvaluation(entity)

if (evaluation_result.variantKey == "on") {
    // feature is on for this entity
} else {
    // feature is off (variantKey is empty, or the flag is disabled)
}
```

### The `enabled` toggle (simplest boolean flag)

The simplest feature flag needs no variants at all. Every flag has a top-level
`enabled` boolean, toggled from the UI's status switch or via
`PUT /api/v1/flags/{flagID}/enabled`:

```bash
curl -X PUT "http://flagr:18000/api/v1/flags/1/enabled" \
  -H 'Content-Type: application/json' \
  -d '{"enabled": false}'
```

When `enabled` is `false`, the evaluator returns a **blank result** immediately —
no `variantKey`, no segment matching, no rollout. It short-circuits before any
segment logic runs. This is the kill switch: flip one boolean, and every
evaluation of that flag returns nothing. No variants or segments to configure.

### The `on`/`off` variant pattern (targeted rollouts)

When you need *targeted* control — on for some users, off for others — use two
variants (`on`/`off`) with a segment and distribution instead of the `enabled`
toggle. The flag stays `enabled: true`; the distribution decides who gets `on`:

```
Variants
  - on
  - off

Segment
  - Constraints (targeted audience, e.g. state == "CA")
  - Rollout Percent: 100%
  - Distribution
    - on: 100%
    - off: 0%
```

![feature flagging setting demo](/images/demo_ff.png)

| Pattern | When to use | How it works |
|---------|-------------|--------------|
| `enabled` toggle | Global kill switch — on or off for everyone | `PUT /flags/{id}/enabled` → evaluator returns blank when `false` |
| `on`/`off` variants | Targeted rollout — on for a specific audience | Flag stays enabled; segment + distribution decide who gets `on` |


## Experimenting — A/B testing

A feature flag answers *on or off?* An experiment answers a harder question:
*which of these is better?* To answer it you expose different variants to
different slices of your audience and measure outcomes. Flagr's job is the
exposure half — deterministic, sticky assignment so the same user always sees
the same treatment. The measurement half (conversion rates, significance) is
yours; Flagr emits the events but does not compute the verdict.

### Variants in experiments: control and treatment

The names `control` and `treatment` are **convention, not a Flagr contract**.
Flagr treats every variant the same — `control` is no more special than `on`,
`green`, or `version_b`. The convention comes from experimentation methodology:

- **Control** — the *baseline* experience. Usually the current production
  behavior. This is what you compare against. If your experiment is "new
  checkout button," control is the existing button.
- **Treatment** — the *change* you are testing. There can be one or many:
  `treatment1`, `treatment2`, `treatment3`. Each is a distinct alternative
  you're measuring against the control.

The distinction matters in your **analytics**, not in Flagr. When you group
exposure rows by `variantKey` in your warehouse, the control group is your
baseline conversion rate; each treatment's rate is compared against it to
measure lift. If you have no control (e.g. testing two brand-new designs),
pick one variant as the reference — but the convention of naming a baseline
"control" makes the analysis self-documenting.

> **Note:** Flagr does not enforce that a flag has a control variant. You can
> run a flag with only treatments, or with no variant named "control" at all.
> The names are for you and your analytics pipeline; Flagr only assigns and
> records the `variantKey` you configured.

To run an A/B test across several variants with a targeted audience, instrument
the code the same way and branch on the assigned variant:

```js
evaluation_result = flagr.postEvaluation(entity)

if (evaluation_result.variantKey == "control") {
    // Control: show the current production checkout (the baseline)
} else if (evaluation_result.variantKey == "treatment1") {
    // Treatment 1: show the new single-page checkout
} else if (evaluation_result.variantKey == "treatment2") {
    // Treatment 2: show the new accordion checkout
} else if (evaluation_result.variantKey == "treatment3") {
    // Treatment 3: show the new one-click checkout
}
```

> **Warning:** Segment order matters. An entity falls into the **first**
> segment whose constraints **all** match. List targeted segments before broad
> ones.

A typical A/B test flag in the UI:

```
Variants
  - control
  - treatment1
  - treatment2
  - treatment3

Segment                         // state == "CA"
  - Constraints (state == "CA")
  - Rollout Percent: 20%
  - Distribution
    - control:    25%
    - treatment1: 25%
    - treatment2: 25%
    - treatment3: 25%

Segment                         // state == "NY" AND age >= 21
  - Constraints (state == "NY" AND age >= 21)
  - Rollout Percent: 100%
  - Distribution
    - control:    50%
    - treatment1:  0%
    - treatment2: 25%
    - treatment3: 25%
```

![ab testing setting demo 1](/images/demo_exp1.png)
![ab testing setting demo 2](/images/demo_exp2.png)

### Measuring experiment outcomes

Assignment from `POST /evaluation` is **not** enough to measure conversion —
you need **impression** events. Use
[Exposure Logging](flagr_exposure.md) when the user actually sees the
treatment, then
[Data Recorders & A/B Analysis](flagr_eval_exposure_pipeline.md) (or your own
consumer on Kafka/Kinesis/Pub/Sub) to build denominators and join to business
metrics. For quick **eval volume** only, see [Datar](flagr_datar.md).

## Dynamic configuration

Sometimes the decision isn't *which experience* but *what value* — the
threshold for a cache, the copy on a button, the timeout for a retry. You
could redeploy for each change, or you can let the variant carry the value.
A variant's **Attachment** is an arbitrary JSON object delivered alongside the
`variantKey`, so your code reads a configuration value the same way it reads
an assignment — no second fetch, no separate config service, no restart.

Use a variant's **Attachment** (an arbitrary JSON object) to carry dynamic
configuration:

```js
evaluation_result = flagr.postEvaluation(entity)
green_color_hex = evaluation_result.variantAttachment["color_hex"]
```

A typical dynamic-config flag:

```
Variants
  - green
    - attachment: {"color_hex": "#42b983"}   // or {"color_hex": "#008000"}
  - red
    - attachment: {"color_hex": "#ff0000"}

Segment
  - Constraints: null
  - Rollout Percent: 100%
  - Distribution
    - green: 100%
    - red: 0%
```

![dynamic configuration demo](/images/demo_dynamic_configuration.png)

> **Note:** Prior to [v1.1.3](https://github.com/openflagr/flagr/releases/tag/1.1.3),
> attachments supported only `string:string` key/value pairs. Current Flagr
> stores attachments as `map[string]any`, so arbitrary JSON values are
> supported. (The historical boundary is documented in the GitHub release
> notes; the in-repo `CHANGELOG.md` links there.)