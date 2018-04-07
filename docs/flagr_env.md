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
