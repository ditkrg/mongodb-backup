# Project Name

## Introduction
This project is a Go application that requires several environment variables to be set for proper configuration. Below is a list of all the available environment variables and their descriptions.

## Environment Variables

| Variable                                | Description                                                | Default   |
| --------------------------------------- | ---------------------------------------------------------- | --------- |
| `MONGODB__CONNECTION_STRING`            | The hostname of the database server.                       | ``        |
| `MONGODB__DB_TO_BACKUP`                 | The port number on which the database server is listening. | ``        |
| `MONGODB__GZIP`                         | The username for connecting to the database.               | `true`    |
| `MONGODB__OPLOG`                        | The username for connecting to the database.               | `false`   |
| `MONGODB__DUMP_DB_USERS_AND_ROLES`      | The password for connecting to the database.               | `true`    |
| `MONGODB__SKIP_USERS_AND_ROLES`         | The password for connecting to the database.               | `false`   |
| `MONGODB__BACKUP_OUT_DIR`               | The API key for accessing external services.               | `/backup` |
| `MONGODB__COLLECTION_TO_BACKUP`         | The API key for accessing external services.               | ``        |
| `MONGODB__EXCLUDED_COLLECTIONS`         | The API key for accessing external services.               | ``        |
| `MONGODB__EXCLUDED_COLLECTION_PREFIXES` | The API key for accessing external services.               | ``        |
| `MONGODB__QUERY`                        | The API key for accessing external services.               | ``        |
| `MONGODB__NUM_PARALLEL_COLLECTIONS`     | The API key for accessing external services.               | `1`       |
| `MONGODB__VERBOSITY__LEVEL`             | The API key for accessing external services.               | `0`       |
| `MONGODB__VERBOSITY__QUIET`             | The API key for accessing external services.               | `false`   |
| `S3__ENDPOINT`                          | The API key for accessing external services.               | ``        |
| `S3__ACCESS_KEY`                        | The API key for accessing external services.               | ``        |
| `S3__SECRET_ACCESS_KEY`                 | The API key for accessing external services.               | ``        |
| `S3__BUCKET`                            | The API key for accessing external services.               | ``        |
| `S3__PREFIX`                            | The API key for accessing external services.               | ``        |
| `S3__KEEP_RECENT_N`                     | The API key for accessing external services.               | `5`       |

