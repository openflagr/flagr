# Server Config

Configuration of Flagr server is derived from the environment variables. Latest [env.go](https://github.com/checkr/flagr/blob/master/pkg/config/env.go).

[env.go](https://raw.githubusercontent.com/checkr/flagr/master/pkg/config/env.go ':include :type=code')

For example

```go
// setting env variable
export FLAGR_DB_DBDRIVER=mysql

// results in
Config.DBDriver = "mysql"
```

## Kinesis Authentication

In order to use Flagr with Kinesis, you need to authenticate with AWS.
For that, you can use the standard AWS authentication methods:

### Environment

The most common way of authentication is over the environemnt, providing the `ACCESS_KEY_ID` and the `SECRET_ACCESS_KEY`. That way flagr can authenticate with AWS to connect to your Kinesis Stream.

e.g.:
```
AWS_ACCESS_KEY_ID=example123
AWS_SECRET_ACCESS_KEY=example123
AWS_DEFAULT_REGION=eu-central-1
```

More info: https://docs.aws.amazon.com/cli/latest/userguide/cli-environment.html

### Other Alternatives

Alternatively, there are couple more options to provide authentication to your stream, such as credentials file, container credentials or instance profiles. Read more about that on the [official AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#config-settings-and-precedence).

**Important**: Make sure the key is attached to a user that has permissions to push records into the stream.
