# PostgreSQL explain plan obfuscation tests

These tests show how explain plans for PostgreSQL should be obfuscated when
SQL obfuscation is enabled. Obfuscation of explain plans for PostgreSQL is
necessary because they can include portions of the original query that may
contain sensitive data.

Each test case consists of a set of files with the following extensions:

* `.query.txt` - the original SQL query that is being explained
* `.explain.txt` - the raw un-obfuscated output from running `EXPLAIN <query>`
* `.colon_obfuscated.txt` - the desired obfuscated explain output if using the
default, more aggressive obfuscation strategy described [here](https://newrelic.atlassian.net/wiki/display/eng/Obfuscating+PostgreSQL+Explain+plans).
* `.obfuscated.txt` - the desired obfuscated explain output if using a more
accurate, less aggressive obfuscation strategy detailed in this
[Jive thread](https://newrelic.jiveon.com/thread/1851).
