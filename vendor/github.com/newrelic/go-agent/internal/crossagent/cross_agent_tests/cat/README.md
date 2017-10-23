### CAT Map test details

The CAT map test cases in `cat_map.json` are meant to be used to verify the
attributes that agents collect and attach to analytics transaction events for
the CAT map project.

**NOTE** currently `nr.apdexPerfZone` is not covered by these tests, make sure you test for this yourself until it is added to these tests.

Each test case should correspond to a simulated transaction in the agent under
test. Here's what the various fields in each test case mean:

| Name | Meaning |
| ---- | ------- |
| `name` | A human-meaningful name for the test case. |
| `appName` | The name of the New Relic application for the simulated transaction. |
| `transactionName` | The final name of the simulated transaction. |
| `transactionGuid` | The GUID of the simulated transaction. |
| `inboundPayload` | The (non-serialized) contents of the `X-NewRelic-Transaction` HTTP request header on the simulated transaction. Note that this value should be serialized to JSON, obfuscated using the CAT obfuscation algorithm, and Base64-encoded before being used in the header value. Note also that the `X-NewRelic-ID` header should be set on the simulated transaction, though its value is not specified in these tests. |
| `expectedIntrinsicFields` | A set of key-value pairs that are expected to be present in the analytics event generated for the simulated transaction. These fields should be present in the first hash of the analytic event payload (built-in agent-supplied fields). |
| `nonExpectedIntrinsicFields` | An array of attribute names that should *not* be present in the analytics event generated for the simulated transaction. |
| `outboundRequests` | An array of objects representing outbound requests that should be made in the context of the simulated transaction. See the table below for details. Only present if the test case involves making outgoing requests from the simulated transaction. |

Here's what the fields of each entry in the `outboundRequests` array mean:

| Name | Meaning |
| ---- | ------- |
| `outboundTxnName` | The name of the simulated transaction at the time this outbound request is made. Your test driver should set the transaction name to this value prior to simulating the outbound request. |
| `expectedOutboundPayload` | The expected (un-obfuscated) content of the outbound `X-NewRelic-Transaction` request header for this request. |
