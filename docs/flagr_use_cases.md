# Flagr Use Cases

**Feature flagging, A/B testing, and dynamic configuration** are all about delivering the experience to the right target audience,
thus they share some components in the product design of Flagr. In fact, Flagr consolidates them together into the concept of
a flag, and the code instrumentation looks similar among them.


## Feature Flagging

A common pattern for feature flagging is a binary on/off toggle. Most of them are kill switches, and sometimes one will have targeted audience of the feature flags. Following is a pseudocode example: given an entity (a user, a request, or a web cookie), Flagr evaluates the entity according to the flag setting.

```
evaluation_result = flagr.post_evaluation( entity )

if (evaluation_result.variant_id == new_feature_on) {
    // do something new and amazing here.
} else {
    // do the current boring stuff.
}
```

And a typical feature flag can be configured from Flagr UI like:

```
Variants
  - on
  - off

Segment
  - Constraints (depends on your targeted audience, e.g. state == "CA")
  - Rollout Percent: 100%
  - Distribution
    - on: 100%
    - off: 0%
```

UI setting example (frontend looks may iterate quickly):
![feature flagging setting demo](/images/demo_ff.png)


## Experimenting - A/B testing

If we want to run A/B testing experiments on several variants with a targeted audience,
we may want to instrument the code to Flagr like the following pseudocode:

```
evaluation_result = flagr.post_evaluation( entity )

if (evaluation_result.variant_id == treatment1) {
    // do the treatment 1 experience
} else if (evaluation_result.variant_id == treatment2) {
    // do the treatment 2 experience
} else if (evaluation_result.variant_id == treatment3) {
    // do the treatment 3 experience
} else {
    // do the control experience
}
```

And a typical A/B testing flag can be configured from Flagr UI like the following:

!> Multiple segments' order is important! Entities will fall
into the **first** segment that match **all** the constraints of it.

```
Variants
  - control
  - treatment1
  - treatment2
  - treatment3

Segment
  - Constraints (state == "CA")
  - Rollout Percent: 20%
  - Distribution
    - control: 25%
    - treatment1: 25%
    - treatment2: 25%
    - treatment3: 25%
Segment
  - Constraints (state == "NY" AND age >= 21)
  - Rollout Percent: 100%
  - Distribution
    - control: 50%
    - treatment1: 0%
    - treatment2: 25%
    - treatment3: 25%
```

UI setting example (frontend looks may iterate quickly):
![ab testing setting demo 1](/images/demo_exp1.png)
![ab testing setting demo 2](/images/demo_exp2.png)


## Dynamic Configuration

One can also leverage the **Variant Attachment** to run dynamic configuration, by supplying a valid JSON object attachment.

!> Before [v1.1.3](https://github.com/openflagr/flagr/releases/tag/1.1.3), only **string:string** key:value pairs were supported inside the JSON object attachment.

For example, the color_hex of green variant can be dynamically configured:

```
evaluation_result = flagr.post_evaluation( entity )
green_color_hex = evaluation_result.variantAttachment["color_hex"]
```

```
Variants
  - green
    - attachment: {"color_hex": "#42b983"} OR {"color_hex": "#008000"}
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
