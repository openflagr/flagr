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

## Pubsub Authentication

You need to authenticate to enable Flagr with Google Cloud Pubsub for data records.
Here's a few ways:

### Gcloud (for development).

```sh
gcloud auth application-default login
```

### Environment

Create and download a service account JSON key and point to it using:

```
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service/account.json
```

> FYI: setting this env var will take over all Google's services on that environment.

The best way to configure service account for Flagr to use pubsub only use:

```
FLAGR_RECORDER_PUBSUB_PROJECT_ID=google-project-id
FLAGR_RECORDER_PUBSUB_KEYFILE=/path/to/service/account.json
```

Basic Authentication for web interface

```
FLAGR_BASIC_AUTH_ENABLED=true
FLAGR_BASIC_AUTH_USERNAME=admin
FLAGR_BASIC_AUTH_PASSWORD=password
```

By default, UI access will prompt for a username/password login. Similar to JWT Auth, prefix and exact paths can be whitelisted to skip the username/password login. The default whitelist will allow api access to `/api/v1/flags` and `/api/v1/evaluation*`

NOTE: this doesn't prevent people from directly curling /api/v1/flags to update flags.

```
FLAGR_BASIC_AUTH_WHITELIST_PATHS="/api/v1/flags,/api/v1/evaluation"
FLAGR_BASIC_AUTH_EXACT_WHITELIST_PATHS=""
```
