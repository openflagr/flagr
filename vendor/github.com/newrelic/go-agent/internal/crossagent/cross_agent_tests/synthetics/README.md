# Synthetics Tests

The Synthetics tests are designed to verify that the agent handles valid and invalid Synthetics requests.

Each test should run a simulated web transaction. A Synthetics HTTP request header is added to the incoming request at the beginning of a web transaction. During the course of the web transaction, an external request is made. And, at the completion of the web transaction, both a Transaction Trace and Transaction Event are recorded.

Each test then verifies that the correct attributes are added to the Transaction Trace and Transaction Event, and the proper request header is added to the external request when required. Or, in the case of an invalid Synthetics request, that the attributes and request header are **not** added.

## Name

| Name | Meaning |
| ---- | ------- |
| `name` | A human-meaningful name for the test case. |

## Settings

The `settings` hash contains a number of key-value pairs that the agent will need to use for configuration for the test.

| Name | Meaning |
| ---- | ------- |
| `agentEncodingKey`| The encoding key used by the agent for deobfuscation of the Synthetics request header. |
| `syntheticsEncodingKey` | The encoding key used by Synthetics to obfuscate the Synthetics request header. In most tests, `agentEncodingKey` and `syntheticsEncodingKey` are the same. |
| `transactionGuid` | The GUID of the simulated transaction. In a non-simulated transaction, this will be randomly generated. But, for testing purposes, you should assign this value as the GUID, since the tests will check for this value to be set in the `nr.guid` attribute of the Transaction Event. |
| `trustedAccountIds` | A list of accounts ids that the agent trusts. If the Synthetics request contains a non-trusted account id, it is an invalid request.|

## Inputs

The input for each test is a Synthetics request header. The test fixture file shows both the de-obfuscated version of the payload, as well as the resulting obfuscated version.

| Name | Meaning |
| ---- | ------- |
| `inputHeaderPayload` | A decoded form of the contents of the `X-NewRelic-Synthetics` request header. |
| `inputObfuscatedHeader` | An obfuscated form of the `X-NewRelic-Synthetics` request header. If you obfuscate `inputHeaderPayload` using the `syntheticsEncodingKey`, this should be the output. |

## Outputs

There are three different outputs that are tested for: Transaction Trace, Transaction Event, and External Request Header.

### outputTransactionTrace

The `outputTransactionTrace` hash contains three objects:

| Name | Meaning |
| ---- | ------- |
| `header` | The last field of the transaction sample array should be set to the Synthetics Resource ID for a Synthetics request, and should be set to `null` if it isn't. (The last field in the array is the 10th element in the header array, but is `header[9]` in zero-based array notation, so the key name is `field_9`.) |
| `expectedIntrinsics` | A set of key-value pairs that represent the attributes that should be set in the intrinsics section of the Transaction Trace. **Note**: If the agent has not implemented the Agent Attributes spec, then the agent should save the attributes in the `Custom` section, and the attribute names should have 'nr.' prepended to them. Read the spec for details. For agents in this situation, they will need to adjust the expected output of the tests accordingly. |
| `nonExpectedIntrinsics` | An array of names that represent the attributes that should **not** be set in the intrinsics section of the Transaction Trace.|

### outputTransactionEvent

The `outputTransactionEvent` hash contains two objects:

| Name | Meaning |
| ---- | ------- |
| `expectedAttributes` | A set of key-value pairs that represent the attributes that should be set in the `Intrinsic` hash of the Transaction Event. |
| `nonExpectedAttributes` | An array of names that represent the attributes that should **not** be set in the `Intrinsic` hash of the Transaction Event. |

### outputExternalRequestHeader

The `outputExternalRequestHeader` hash contains two objects:

| Name | Meaning |
| ---- | ------- |
| `expectedHeader` | The outbound header that should be added to external requests (similar to the CAT header), when the original request was made from a valid Synthetics request. |
| `nonExpectedHeader` | The outbound header that should **not** be added to external requests, when the original request was made from a non-Synthetics request. |
