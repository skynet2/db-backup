## DB-BACKUP Tool

![build workflow](https://github.com/skynet2/db-backup/actions/workflows/release.yaml/badge.svg?branch=master)
[![go-report](https://goreportcard.com/badge/github.com/skynet2/db-backup)](https://goreportcard.com/report/github.com/skynet2/db-backup)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/skynet2/db-backup)](https://pkg.go.dev/github.com/skynet2/db-backup?tab=doc)

### Simple database backup tool with support for remote storage.

## Supported databases (providers)
- [x] postgres (requires pg_dump)
- [ ] mssql

## Supported storages (providers)
- [x] s3
- [x] s3 compatible (provider = s3)
- [ ] ftp

## Supported notification channels (providers)
- [x] discord
- [x] telegram
- [x] mattermost
- [ ] slack

## Configuration example
```yml
exclude_dbs:
  - stats
db:
  provider: postgres
  dump_dir: "/var/backup/"
  postgres:
    host: "127.0.0.1"
    port: 5432
    user: postgres
    password: qwerty
    db_default_name: postgres
    tls_enabled: false
storage:
  provider: s3
  dir_template: "{{.Host}}/{{.DbName}}"
  max_files: 3
  s3:
    region: "nl-ams"
    endpoint: "https://s3.nl-ams.scw.cloud"
    bucket: "myawesomebacket"
    access_key: "access_key"
    secret_key: "secret_key"
    disable_ssl: true
    force_path_style: false
notifications:
  success:
    channels:
      - type: telegram
        token: "bot_token"
        chat: "chat_id"
      - type: discord
        webhook: "https://my-discord-webhook-url"
```
### Configuration fields
* include_dbs - include only specified databases in backup job
* exclude_dbs - exclude specific databases from backup job
* db - database connection settings
  * provider - database provider (ex. postgres)
  * dump_dir - temporary directory for backup process
  * postgres - postgres provider configuration
    * host - server ip\hostname
    * port - port
    * user - user
    * password - password
    * db_default_name - default database name (postgres)
    * tls_enabled - tls configuration (true\false)
    * compression_level = database compression level (pg_dump configuration), 5 by default
* storage
  * provider - storage provider (ex. s3)
  * dir_template - golang template for remote directory. supported values : {{.Host}} and {{.DbName}}
  * max_files - max remote backups for specific database. For example max_files = 5 and if we already have 5 files for that database at remote storage, the oldest file will be removed
  * s3 - s3 provider configuration
    * region - region
    * endpoint - endpoint (s3 compatible storages)
    * bucket - bucket
    * access_key - access_key
    * secret_key - secret_key
    * disable_ssl - disable_ssl (true\false)
    * force_path_style - force_path_style (true\false)
* Notifications
  * success - will be called on success 
    * channels - array of notification channels
      * type - channel type (ex. discord or telegram)
      * token - required for telegram (bot token)
      * chat - chat_id (telegram)
      * webhook - webhook url (discord)
    * template - custom go template
  * fail - exactly same as success, but will be executed on fail or error. if fail - empty, success will be used
