## Datastore instance tests

The datastore instance tests provide attributes similar to what an agent could expect to find regarding a database configuration and specifies the expected [datastore instance metric](https://source.datanerd.us/agents/agent-specs/blob/master/Datastore-Metrics-PORTED.md#datastore-metric-namespace) that should be generated. The table below lists types attributes and whether will will always be included or optionally included in each test case.

| Name | Present | Description |
|---|---|---|
| system_hostname | always | the hostname of the machine |
| db_hostname | sometimes | the hostname reported by the database adapter |
| product | always | the database product for this configuration
| port | sometimes | the port reported by the database adapter |
| unix_socket | sometimes |the path to a unix domain socket reported by a database adapter |
| database_path | sometimes |the path to a filesystem database |
| expected\_instance\_metric | always | the instance metric expected to be generated from the given attributes |

## Implementing the test cases
The idea behind these test cases are that you are able to determine a set of configuration properties from a database connection, and based on those properties you should generate the `expected_instance_metric`. Sometimes the properties available are minimal and will mean that you will need to fall back to defaults to obtain some of the information. When there is missing information from a database adapter the guiding principle is to fill in the defaults when they can be inferred, but do not make guesses that could be incorrect or misleading. Some agents may have access to better data and may not need to make inferences. If this applies to your agent then many of these tests will not be applicable.
