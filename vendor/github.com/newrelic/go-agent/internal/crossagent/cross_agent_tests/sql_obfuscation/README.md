These test cases cover obfuscation (more properly, masking) of literal values
from SQL statements captured by agents. SQL statements may be captured and
attached to transaction trace nodes, or to slow SQL traces.

`sql_obfuscation.json` contains an array of test cases.  The inputs for each
test case are in the `sql` property of each object. Each test case also has an
`obfuscated` property which is an array containing at least one valid output.

Test cases also have a `dialects` property, which is an array of strings which
specify which sql dialects the test should apply to. See "SQL Syntax Documentation" list below. This is relevant because for example, PostgreSQL uses
different identifier and string quoting rules than MySQL (most notably,
double-quoted string literals are not allowed in PostgreSQL, where
double-quotes are instead used around identifiers).

Test cases may also contain the following properties:
  * `malformed`: (boolean) tests who's SQL queries are not valid SQL in any
  quoting mode. Some agents may choose to attempt to obfuscate these cases,
  and others may instead just replace the query entirely with a placeholder
  message.
  * `pathological`: (boolean) tests which are designed specifically to break
  specific methods of obfuscation, or contain patterns that are known to be
  difficult to handle correctly
  * `comments`: an array of strings that could be usefult for understanding
  the test.

The following database documentation may be helpful in understanding these test
cases:
* [MySQL String Literals](http://dev.mysql.com/doc/refman/5.5/en/string-literals.html)
* [PostgreSQL String Constants](http://www.postgresql.org/docs/8.2/static/sql-syntax-lexical.html#SQL-SYNTAX-CONSTANTS)

SQL Syntax Documentation:
* [MySQL](http://dev.mysql.com/doc/refman/5.5/en/language-structure.html)
* [PostgreSQL](http://www.postgresql.org/docs/8.4/static/sql-syntax.html)
* [Cassandra](http://docs.datastax.com/en/cql/3.1/cql/cql_reference/cql_lexicon_c.html)
* [Oracle](http://docs.oracle.com/cd/B28359_01/appdev.111/b28370/langelems.htm)
* [SQLite](https://www.sqlite.org/lang.html)
