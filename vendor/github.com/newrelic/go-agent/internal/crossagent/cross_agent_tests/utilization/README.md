# The Utilization Tests

The Utilization tests ensure that the appropriate information is being gathered for pricing. It is centered around ensuring the JSON is correct. Each JSON block is a test case, with potentially the following fields:

  - testname: The name of the test
  - input_total_ram_mib: The total ram number calculated by the agent.
  - input_logical_processors: The number of logical processors calculated by the agent.
  - input_hostname: The hostname calculated by the agent.
  - input_aws_id: The aws id determined by the agent.
  - input_aws_type: The aws type determined by the agent.
  - input_aws_zone: The aws zone determined by the agent.
  - input_environment_variables: Any environment variables which have been set.
  - expected_output_json: The expected JSON output from the agent for the utilization hash.
