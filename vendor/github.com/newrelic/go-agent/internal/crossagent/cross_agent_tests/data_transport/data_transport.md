What
----

The Data Transport Tests are meant to validate your agent's data transport layer (the part of the code that sends requests to the collector and interprets the responses). These tests are meant to be consumable both by code (for automated tests) and by humans, such that we can treat these tests as a sort of spec for agent-collector communication.

The basic gist is that each of these tests are just collections of steps. A step is one of the following:

 * An event that you need to induce your agent to do (such as generate a metric or do a harvest).
 * An expectation that your agent should send a request (as a result of previous steps).
 * A composite step, which is just several other steps grouped together to reduce repetition.

### Types of steps (not an exhaustive list) ###

 * `event_agent_start` -- Represents the startup of the agent. Will contain a `payload` property that defines the startup configuration.
 * `event_metric` -- Represents some event that would cause your agent to generate a metric (such as a page view, a database query, etc).
 * `event_harvest_metrics` -- Represents some event that would cause your agent to harvest metrics (such as a harvest timer elapsing).
 * `event_local_config_update` -- Represents a change to local config while the agent is running.
 * `expect_request` -- Represents an expectation that a particular request should happen as a result of the previous events.
	 * Will sometimes contain a `payload` property that defines the expected serialized payload (if omitted, payload can be ignored).
		 * Expected payloads will sometimes contain wildcard tokens such as `"__ANY_FLOAT__"`.
	 * Will sometimes contain a `response_payload` property that defines the response that the test runner should give to the agent (if omitted, just send back any payload that makes your agent continue on happily).
 * `expect_no_request` -- Represents an expectation that **no** request should happen as a result of the previous events. *Note that this expectation is redundant if your test runner follows that paradigm that every request that occurs must be explicitly called out by the test data.*

### "But our agent doesn't do *xyz*!" ###

It is inevitable that there will be conflicts in functionality between the various agents. As much as possible, these tests are written to be idealistically comprehensive -- that is, covering all behavior that a perfectly functioning agent should follow -- but flexible enough for agents to intentionally ignore components that either don't apply or are not yet supported.

Examples:

 * **Discrepancy:** Agent does not send `agent_settings` command.
	 * **Solution:** Ignore `expect_request` steps with `"agent_settings"` as the command name.
 * **Discrepancy:** Agent does not yet support custom events.
	 * **Solution:** Ignore any test that contains a `event_custom_event` step.
 * **Discrepancy:** Agent request payloads looks significantly different than the expected payload due to special reasons X, Y, and Z.
	 * **Solution:** Pre-process all expected payloads to make them match your agent's goofy output.