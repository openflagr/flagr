# Migrations

Users can create flag collections as yaml files within a migration directory and run flagr to insert/modify flags. 
This allows for developers to create flags as deployment assets and allows for migrating flags through various environments whilst keeping keys stable.

Each found flag is upserted into the database. Then the flag has all segments, variants and tags removed then replaced with the new flag properties.

As an example, create a file named `migrations/202403030000.yaml` with the following content:
```yaml
---
# this is a basic flag
- key: SIMPLE-FLAG-1
  description: a toggle for just one user
  enabled: true
  segments:
    - description: flag for just for one email test@test.com
      rank: 0
      rolloutPercent: 100
      constraints:
        - property: email
          operator: EQ
          value: '"test@test.com"'
      distributions:
        - variantKey: "on"
          percent: 100
    - rank: 1
      rolloutPercent: 100
      constraints: []
      distributions:
        - variantKey: "off"
          percent: 100
  variants:
    - key: "off"
      attachment: {}
    - key: "on"
      attachment: {}
  entityType: User
  dataRecordsEnabled: true

```

```shell
$ flagr -m 
INFO[0146] 1 new migrations completed (1 total) 
```
Once the application is ran, flagr will scan the migration files, insert them into the db and shut down.

### Config
Location of yaml configs can be set with either argument or env var.
```
FLAGR_MIGRATION_PATH=./migrations/ ./flagr -m
./flagr -m --migrationPath=`pwd`/migrations
```
