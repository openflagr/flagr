# Cross Agent Tests

### Data Policy

None of these tests should contain customer data such as SQL strings.
Please be careful when adding new tests from real world failures.

### Access

Push access to this repository is granted via membership in the cross-agent-team GHE group. Contact Belinda Runkle if you are on the agent team but don't have push access.

### Tests

| Test Files    | Description   |
| ------------- |-------------|
| [rum_loader_insertion_location](rum_loader_insertion_location) | Describe where the RUM loader (formerly known as header) should be inserted. |
| [rum_footer_insertion_location](rum_footer_insertion_location) | Describe where the RUM footer (aka "client config") should be inserted.  These tests do not apply to agents which insert the footer directly after the loader. |
| [rules.json](rules.json) | Describe how url/metric/txn-name rules should be applied. |
| [rum_client_config.json](rum_client_config.json) | These tests dictate the format and contents of the browser monitoring client configuration.  For more information see: [SPEC](https://newrelic.atlassian.net/wiki/display/eng/BAM+Agent+Auto-Instrumentation) |
| [sql_parsing.json](sql_parsing.json) | These tests show how an SQL string should be parsed for the operation and table name. |
| [url_clean.json](url_clean.json) | These tests show how URLs should be cleaned before putting them into a trace segment's parameter hash (under the key 'uri'). |
| [url_domain_extraction.json](url_domain_extraction.json) | These tests show how the domain of a URL should be extracted (for the purpose of creating external metrics). |
| [postgres_explain_obfuscation](postgres_explain_obfuscation) | These tests show how plain-text explain plan output from PostgreSQL should be obfuscated when SQL obfuscation is enabled. |
| [sql_obfuscation](sql_obfuscation) | Describe how agents should obfuscate SQL queries before transmission to the collector. |
| [attribute_configuration](attribute_configuration.json) | These tests show how agents should respond to the various attribute configuration settings.  For more information see: [Attributes SPEC](https://source.datanerd.us/agents/agent-specs/blob/master/Agent-Attributes-PORTED.md) |
| [cat](cat) | These tests cover the new Dirac attributes that are added for the CAT Map project. See the [CAT Spec](https://source.datanerd.us/agents/agent-specs/blob/master/Cross-Application-Tracing-PORTED.md) and the [README](cat/README.md) for details.|
| [labels](labels.json) | These tests cover the Labels for Language Agents project. See the [Labels for Language Agents Spec](https://newrelic.atlassian.net/wiki/display/eng/Labels+for+Language+Agents) for details.|
| [proc_cpuinfo](proc_cpuinfo) | These test correct processing of `/proc/cpuinfo` output on Linux hosts. |
| [proc_meminfo](proc_meminfo) | These test correct processing of `/proc/meminfo` output on Linux hosts. |
| [transaction_segment_terms.json](transaction_segment_terms.json) | These tests cover agent implementations of the `transaction_segment_terms` transaction renaming rules introduced in collector protocol 14. See [the spec](https://newrelic.atlassian.net/wiki/display/eng/Language+agent+transaction+segment+terms+rules) for details. |
| [synthetics](synthetics) | These tests cover agent support for Synthetics. For details, see [Agent Support for Synthetics: Forced Transaction Traces and Analytic Events](https://source.datanerd.us/agents/agent-specs/blob/master/Synthetics-PORTED.md). |
| [docker_container_id](docker_container_id) | These tests cover parsing of Docker container IDs from `/proc/*/cgroup` on Linux hosts. |
| [utilization](utilization) | These tests cover the collection and validation of metadata for billing purposes as per the [Utilization spec](https://source.datanerd.us/agents/agent-specs/blob/master/Utilization.md). |
| [utilization_vendor_specific](utilization_vendor_specific) | These tests cover the collection and validation of metadata for AWS, Pivotal Cloud Foundry, Google Cloud Platform, and Azure as per the [Utilization spec](https://source.datanerd.us/agents/agent-specs/blob/master/Utilization.md). |
| [data_transport](data_transport/data_transport.json) | These test correct JSON payload generation and handling collector responses. See [readme](https://source.datanerd.us/agents/cross_agent_tests/blob/master/data_transport/data_transport.md) for details. |
