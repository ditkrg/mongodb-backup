# Project Name

## Introduction
This project is a Go application that requires several environment variables to be set for proper configuration. Below is a list of all the available environment variables and their descriptions.

## Environment Variables

| Variable                                | Description                                                | Default   |
| --------------------------------------- | ---------------------------------------------------------- | --------- |
| `MONGODB__CONNECTION_STRING`            | The hostname of the database server.                       | ``        |
| `MONGODB__DB_TO_BACKUP`                 | The port number on which the database server is listening. | ``        |
| `MONGODB__GZIP`                         | should the back up be compressed into gzip                 | `true`    |
| `MONGODB__OPLOG`                        |                                                            | `false`   |
| `MONGODB__DUMP_DB_USERS_AND_ROLES`      |                                                            | `true`    |
| `MONGODB__SKIP_USERS_AND_ROLES`         |                                                            | `false`   |
| `MONGODB__BACKUP_OUT_DIR`               | The location where to output the backup                    | `/backup` |
| `MONGODB__COLLECTION_TO_BACKUP`         | Name of a collection to backup                             | ``        |
| `MONGODB__EXCLUDED_COLLECTIONS`         | Name of collection to exclude                              | ``        |
| `MONGODB__EXCLUDED_COLLECTION_PREFIXES` | Prefix for collections to exclude                          | ``        |
| `MONGODB__QUERY`                        |                                                            | ``        |
| `MONGODB__NUM_PARALLEL_COLLECTIONS`     | Number of collections to backup in parallel                | `1`       |
| `MONGODB__VERBOSITY__LEVEL`             | The log level                                              | `0`       |
| `MONGODB__VERBOSITY__QUIET`             |                                                            | `false`   |
| `S3__ENDPOINT`                          | The s3 endpoint to upload the backups to                   | ``        |
| `S3__ACCESS_KEY`                        | The access key                                             | ``        |
| `S3__SECRET_ACCESS_KEY`                 | The secret access key                                      | ``        |
| `S3__BUCKET`                            | The bucket name to upload to                               | ``        |
| `S3__PREFIX`                            | The prefix of the backups (folder for the backups)         | ``        |
| `S3__KEEP_RECENT_N`                     | The number of the backups to retain                        | `5`       |

