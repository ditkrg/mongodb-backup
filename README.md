# MongoDB Backup CLI Tool

A Golang-powered CLI tool for managing MongoDB backups with seamless integration to S3.

## Table of Contents
- [MongoDB Backup CLI Tool](#mongodb-backup-cli-tool)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
    - [Option 1: Download Pre-built Binary (Recommended)](#option-1-download-pre-built-binary-recommended)
    - [Option 2: Build from Source](#option-2-build-from-source)
  - [Commands and Flags](#commands-and-flags)
    - [Global Flags](#global-flags)
    - [Commands](#commands)
      - [1. **`list`**: List backups](#1-list-list-backups)
      - [2. **`dump`**: Take a database or point-in-time backup](#2-dump-take-a-database-or-point-in-time-backup)
      - [3. **`restore`**: Restore a database/point-in-time backup](#3-restore-restore-a-databasepoint-in-time-backup)
  - [Examples](#examples)
    - [Basic Usage](#basic-usage)
    - [Using Environment Variables](#using-environment-variables)
  - [Permissions](#permissions)
    - [Backup \& Oplog Backup](#backup--oplog-backup)
    - [Restore](#restore)
    - [Oplog Restore](#oplog-restore)
  - [Troubleshooting](#troubleshooting)
    - [Getting Help](#getting-help)
  - [Important Behaviors and Prerequisites](#important-behaviors-and-prerequisites)
    - [Backup Dependencies](#backup-dependencies)
    - [Restore Behaviors](#restore-behaviors)
    - [Backup Retention](#backup-retention)
  - [⚠️ Warning: User and Role Restoration Behavior](#️-warning-user-and-role-restoration-behavior)
  - [Contribution](#contribution)

## Features
- **Full Database Backup & Restore**: Create complete backups of your MongoDB databases
- **Oplog Backup & Restore**: Perform point-in-time backups and restores using oplog
- **S3 Integration**: Seamlessly store and manage backups in S3-compatible storage
- **Flexible Configuration**: Support for environment variables and command-line flags
- **Cross-Platform**: Available for Linux, Windows, and macOS (Intel & Apple Silicon)
- **Docker Support**: Ready-to-use Docker image for containerized environments

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Download the latest pre-built binary for your platform from our [releases page](https://github.com/ditkrg/mongodb-backup/releases).

1. Choose the appropriate binary for your platform:
   - Linux: `mongodb-backup-linux-amd64`
   - Windows: `mongodb-backup-windows-amd64.exe`
   - macOS Intel: `mongodb-backup-darwin-amd64`
   - macOS Apple Silicon: `mongodb-backup-darwin-arm64`

2. Make the binary executable (Linux/macOS):
   ```bash
   chmod +x mongodb-backup-*
   ```

3. Add to PATH:
   - **Linux/macOS**:
     ```bash
     sudo mv mongodb-backup-* /usr/local/bin/mongodb-backup
     ```
   - **Windows**:
     - Move the `.exe` file to a directory of your choice
     - Add that directory to your system's PATH environment variable
     - Or use the binary directly from the download location

### Option 2: Build from Source

If you prefer to build from source:

1. Clone the repository:
   ```bash
   git clone https://github.com/ditkrg/mongodb-backup.git
   cd mongodb-backup
   ```

2. Build the binary:

   - For Ubuntu/Linux
     ```bash
     GOOS=linux GOARCH=amd64 go build -o mongodb-backup
     ```

   - For Windows
     ```bash
     GOOS=windows GOARCH=amd64 go build -o mongodb-backup.exe
     ```

   - For macOS
     ```bash
     # For Intel Macs
     GOOS=darwin GOARCH=amd64 go build -o mongodb-backup

     # For Apple Silicon Macs
     GOOS=darwin GOARCH=arm64 go build -o mongodb-backup
     ```

3. Add the binary to your PATH:

   - For Ubuntu/Linux
     ```bash
     sudo mv mongodb-backup /usr/local/bin
     ```

   - For Windows
     - Move the `mongodb-backup.exe` to a directory of your choice
     - Add that directory to your system's PATH environment variable
     - Or use the binary directly from the build location

   - For macOS
     ```bash
     sudo mv mongodb-backup /usr/local/bin
     ```


## Commands and Flags

### Global Flags

These flags apply across all commands:

- `-h, --help`: Show context-sensitive help.
- `-v, --version`: Print the version number.

### Commands

#### 1. **`list`**: List backups
Lists the available backups stored in S3, with options to filter by type or database.

**Usage**:
```bash
mongodb-backup list --s3-endpoint=STRING --s3-access-key=STRING --s3-secret-key=STRING --s3-bucket=STRING [flags]
```

**Flags**:
- `--oplog`: List only oplog backups.
- `--full-backups`: List only full backups.
- `--database=STRING`: list backups for a given database (full backups are not included).

**S3 Flags**:
- `--s3-endpoint=STRING ($S3__ENDPOINT)`: S3 endpoint.
- `--s3-access-key=STRING ($S3__ACCESS_KEY)`: S3 access key.
- `--s3-secret-key=STRING ($S3__SECRET_ACCESS_KEY)`: S3 secret access key.
- `--s3-bucket=STRING ($S3__BUCKET)`: S3 bucket name.
- `--s3-prefix=STRING ($S3__PREFIX)`: (Optional) S3 path prefix.

**Verbosity Flags**:
- `--verbosity-level=1 ($VERBOSITY__LEVEL)`: Log verbosity level (1-3, higher is more verbose).
- `--verbosity-quiet ($VERBOSITY__QUIET)`: Suppress all log output.


#### 2. **`dump`**: Take a database or point-in-time backup
Creates a MongoDB backup and uploads it to S3.

**Usage**:
```bash
main dump --s3-endpoint=STRING --s3-access-key=STRING --s3-secret-key=STRING --s3-bucket=STRING --connection-string=STRING [flags]
```

**S3 Flags**:
- `--s3-endpoint=STRING ($S3__ENDPOINT)`: S3 endpoint.
- `--s3-access-key=STRING ($S3__ACCESS_KEY)`: S3 access key.
- `--s3-secret-key=STRING ($S3__SECRET_ACCESS_KEY)`: S3 secret access key.
- `--s3-bucket=STRING ($S3__BUCKET)`: S3 bucket name.
- `--s3-prefix=STRING ($S3__PREFIX)`: (Optional) S3 path prefix.


**Common Mongo Dump Flags**:
- `--connection-string=STRING ($MONGO_DUMP__CONNECTION_STRING)`: MongoDB URI.
- `--backup-dir=STRING ($MONGO_DUMP__BACKUP_DIR)`: Directory to store the backup locally (default: `/backup`).
- `--keep-recent-n=NUMBER ($MONGO_DUMP__KEEP_RECENT_N)`: Number of backups to keep.

**Namespace Options**:
- `--database=STRING ($MONGO_DUMP__DATABASE)`: (Optional) Database to back up.
- `--collection=STRING ($MONGO_DUMP__COLLECTION)`: (Optional) Collection to back up.

**Query Options**:
- `--query=STRING ($MONGO_DUMP__QUERY)`: (Optional) Query filter in JSON format.
- `--query-file=STRING ($MONGO_DUMP__QUERY_FILE)`: (Optional) Path to a file containing the query filter.
- `--read-preference=STRING ($MONGO_DUMP__READ_PREFERENCE)`: (Optional) Read preference (e.g., `nearest`).

**Output Options**:
- `--gzip/--no-gzip ($MONGO_DUMP__GZIP)`: Compress the backup with gzip.
- `--oplog ($MONGO_DUMP__OPLOG)`: Take an oplog backup instead of database backup.
- `--dump-db-users-and-roles ($MONGO_DUMP__DUMP_DB_USERS_AND_ROLES)`: Dump the users and roles in the databas
- `--skip-users-and-roles ($MONGO_DUMP__SKIP_USERS_AND_ROLES)`: Skip dumping the users and roles in the database
- `--excluded-collections=COLLECTIONS,... ($MONGO_DUMP__EXCLUDED_COLLECTIONS)`: (Optional) Collections to exclude from the backup.
- `--excluded-collection-prefixes=PREFIXES,... ($MONGO_DUMP__EXCLUDED_COLLECTION_PREFIXES)`: (Optional) Prefixes of collections to exclude.
- `--num-parallel-collections=N ($MONGO_DUMP__NUM_PARALLEL_COLLECTIONS)`: The number of collections to dump in parallel

**Verbosity Options**:
- `--verbosity-level=1 ($VERBOSITY__LEVEL)`: Log verbosity level (1-3, higher is more verbose).
- `--verbosity-quiet ($VERBOSITY__QUIET)`: Suppress all log output.


#### 3. **`restore`**: Restore a database/point-in-time backup
Restores a MongoDB database from a backup stored in S3.

**Usage**:
```bash
mongodb-backup restore --s3-endpoint=STRING --s3-access-key=STRING --s3-secret-key=STRING --s3-bucket=STRING --connection-string=STRING --backup-dir=STRING [flags]
```

**Common Mongo Dump Flags**:
- `--connection-string=STRING ($MONGO_RESTORE__CONNECTION_STRING)`: MongoDB URI.
- `--backup-dir=STRING ($MONGO_RESTORE__BACKUP_DIR)`: Directory to store the backup locally (default: `/backup`).

**S3 Flags**:
- `--s3-key=STRING ($S3__KEY)`:  The key of the backup to restore (Include the bucket, prefix, and key name in path style `bucket/prefix/key`).
- `--s3-endpoint=STRING ($S3__ENDPOINT)`: S3 endpoint.
- `--s3-access-key=STRING ($S3__ACCESS_KEY)`: S3 access key.
- `--s3-secret-key=STRING ($S3__SECRET_ACCESS_KEY)`: S3 secret access key.
- `--s3-bucket=STRING ($S3__BUCKET)`: S3 bucket name.
- `--s3-prefix=STRING ($S3__PREFIX)`: (Optional) S3 path prefix.

**Namespace Options**:
- `--database=STRING ($MONGO_RESTORE__DATABASE)`: Database to restore.
- `--collection=STRING ($MONGO_RESTORE__COLLECTION)`: Collection to restore.
- `--ns-include=INCLUDES,... ($MONGO_RESTORE__NS_INCLUDE)`: Namespaces (database.collection) to include.
- `--ns-exclude=EXCLUDES,... ($MONGO_RESTORE__NS_EXCLUDE)`: Namespaces to exclude.

**Restore Options**:
- `--users-to-skip-disable ($USERS_TO_SKIP_DISABLE)` List of users to skip disabling, make sure to provide the admin user and the user that will be used to restore the backup, it has to the ***users ID*** ( it is usually in the following format `database.username`), you can get the users ID by running the following command in the MongoDB shell: `db.getUsers()`
- `--drop ($MONGO_RESTORE__DROP)`: Drop each collection before import
- `--dry-run ($MONGO_RESTORE__DRY_RUN)`: Run the restore in `dry run` mode
- `--write-concern="majority ($MONGO_RESTORE__WRITE_CONCERN `: Write concern for the restore operation
- `--no-index-restore ($MONGO_RESTORE__NO_INDEX_RESTORE)`: Don't restore indexes
- `--convert-legacy-indexes ($MONGO_RESTORE__CONVERT_LEGACY_INDEXES)`: Removes invalid index options and rewrites legacy option values
- `--no-options-restore ($MONGO_RESTORE__NO_OPTIONS_RESTORE)`: Don't restore collection options
- `--[no-]keep-index-version ($MONGO_RESTORE__KEEP_INDEX_VERSION)`: Don't upgrade indexes to latest version
- `--maintain-insertion-order ($MONGO_RESTORE__MAINTAIN_INSERTION_ORDER)`: restore the documents in the order of their appearance in the input source. By default the insertions will be performed in an arbitrary order. Setting this flag also enables the behavior of stopOnError and restricts NumInsertionWorkersPerCollection to 1.
- `--num-parallel-collections=1 ($MONGO_RESTORE__NUM_PARALLEL_COLLECTIONS)`: Number of collections to restore in parallel
- `--num-insertion-workers=1 ($MONGO_RESTORE__NUM_INSERTION_WORKERS)`: Number of insert operations to run concurrently per collection
- `--stop-on-error ($MONGO_RESTORE__STOP_ON_ERROR)`: Stop restoring at first error rather than continuing
- `--bypass-document-validation ($MONGO_RESTORE__BYPASS_DOCUMENT_VALIDATION)`: Bypass document validation
- `--preserve-uuid ($MONGO_RESTORE__PRESERVE_UUID)`: preserve original collection UUIDs (requires drop)
- `--fix-dotted-hashed-indexes ($MONGO_RESTORE__FIX_DOTTED_HASHED_INDEXES)`: when enabled, all the hashed indexes on dotted fields will be created as single field ascending indexes on the destination
- `--[no-]object-check ($MONGO_RESTORE__OBJECT_CHECK)`: validate all objects before inserting
- `--[no-]oplog-replay ($MONGO_RESTORE__OPLOG_REPLAY)`: replay the oplog backups
- `--oplog-limit=STRING ($MONGO_RESTORE__OPLOG_LIMIT_TO)`: The End time of the OpLog restore
- `--[no-]gzip ($MONGO_RESTORE__GZIP)`: Whether the backup is gzipped
- `--restore-db-users-and-roles ($MONGO_RESTORE__RESTORE_DB_USERS_AND_ROLES)`: restore user and role definitions for the given database
- `--skip-users-and-roles ($MONGO_RESTORE__SKIP_USERS_AND_ROLES)`: Skip restoring users and roles, regardless of namespace, when true

**Verbosity Options**:
- `--verbosity-level=1`: Log verbosity level (1-3, higher is more verbose).
- `--verbosity-quiet`: Suppress all log output.


## Examples

### Basic Usage

1. **List All Backups**
   ```bash
   mongodb-backup list \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --oplog|--full-backups|--database="your-database"
   ```

2. **Create a Full Backup**
   ```bash
   mongodb-backup dump \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
   ```
3. **Create a Database Backup of a specific database**
   ```bash
   mongodb-backup dump \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
     --database="mydb"
   ```


4. **Create Oplog Backup** (You need to have a full backup in the S3 bucket first)
   ```bash
   mongodb-backup dump \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
     --oplog
   ```

5. **Restore a Database Backup**
   ```bash
   mongodb-backup restore \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
     --s3-key="key-of-the-database-backup"
   ```

6. **Restore a full backup** (by default it will restore the full backup and then start replaying the oplog from the time of the chosen backup)
   ```bash
   mongodb-backup restore \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
     --s3-key="key-of-the-full-backup"
   ```

7. **Restore specific database from a full backup**
   ```bash
   mongodb-backup restore \
     --s3-endpoint="https://s3.example.com" \
     --s3-access-key="your-access-key" \
     --s3-secret-key="your-secret-key" \
     --s3-bucket="your-backups" \
     --connection-string="mongodb://localhost:27017" \
     --database="mydb"|--ns-exclude="mydb.*"
   ```

### Using Environment Variables

You can use environment variables instead of command-line flags:

```bash
export S3__ENDPOINT="https://s3.example.com"
export S3__ACCESS_KEY="your-access-key"
export S3__SECRET_ACCESS_KEY="your-secret-key"
export S3__BUCKET="your-backups"
export MONGO_DUMP__CONNECTION_STRING="mongodb://localhost:27017"

mongodb-backup dump --database="mydb"
```

or set the environment variables in the `.env` file then run the command:

```bash
export $(cat .env)
mongodb-backup dump --database="mydb"
```

## Permissions

### Backup & Oplog Backup

To successfully perform a `dump` operation using this CLI tool, ensure the MongoDB user has the necessary permissions, the user must have the `backup` role or equivalent permissions on all databases, including the `admin` database. This allows the user to:
   - Read data from all collections.
   - Access system collections required for metadata.

Example command to create a user with the required permissions:
```bash
db.createUser({
   user: "<username>",
   pwd: "<password>",
   roles: [
      { role: "backup", db: "admin" }
   ]
})
```

Failing to configure these permissions may result in errors or incomplete backups. Always validate user roles before performing a `dump` operation.

### Restore

To successfully perform a `restore` operation using this CLI tool, ensure the MongoDB user has the necessary permissions, the user must have the `restore` role or equivalent permissions on all databases, including the `admin` database. This allows the user to:
  - Create collections.
  - Insert documents.
  - Restore indexes, users, and roles.

Example command to create a user with the required permissions:
```bash
db.createUser({
   user: "<username>",
   pwd: "<password>",
   roles: [
       { role: "restore", db: "admin" }
   ]
})
```

Failing to configure these permissions may result in incomplete or unsuccessful restores. Always validate user roles before performing a `restore` operation.


### Oplog Restore
To perform an oplog restore, the user must have a specific role with the following privileges:

```json
{
  "actions": ["anyAction"],
  "resource": { "anyResource": true }
}
```

use can create the role by running the following commands in the MongoDB shell:

```bash
use admin

db.createRole({
  role: "myroot",
  privileges: [{ actions: ["anyAction"], resource: { anyResource: true } }],
  roles: []
})
```
then you can create the user with the role by running the following commands in the MongoDB shell:

```bash
db.adminCommand({
  createUser: "<Username>",
  pwd: "<Password>",
  roles: [
    { role: "root", db: "admin" },
    { role: "myroot", db: "admin" }
  ]
})
```

## Troubleshooting

### Getting Help
- Open an issue on GitHub for bug reports
- Check existing issues for known problems
- Include logs with `--verbosity-level=3` for better debugging

## Important Behaviors and Prerequisites

### Backup Dependencies

1. **Full Backup**
   - A full backup is the foundation for all other backup operations
   - Full backups include all databases, collections, and optionally user/role definitions
   - Required before taking oplog backups

2. **Oplog Backup Requirements**
   - Must have at least one full backup in the S3 bucket before taking oplog backups
   - Each oplog backup tracks changes since the last backup Oplog backup, except for the first oplog backup, it will track changes since the full backup
   - Multiple oplog backups can be taken after a full backup

### Restore Behaviors

1. **Full Backup Restore**
   - When restoring a full backup, the tool will:
     - First restore the full backup data
     - Automatically replay all available oplog entries from the backup time
     - This provides point-in-time recovery without manual oplog restoration

2. **Database-Specific Restore**
   - When restoring a specific database:
     - Can restore from either a full backup or database-specific backup
     - If restoring from a full backup, only the specified database is restored
     - Oplog entries ***will not*** be replayed for the specified database.

3. **Oplog Restore**
   - Oplog restore is automatic when restoring a full backup
   - The tool will:
     1. Restore the full backup
     2. Find all oplog backups taken after the full backup
     3. Apply oplog entries in chronological order
   - You cannot manually trigger oplog restore; it's part of the full restore process

### Backup Retention

1. **Keep Recent N Backups**
   - Use `--keep-recent-n` to maintain a specific number of backups
   - Applies separately to:
     - Full backups
     - Database-specific backups
   - Older backups are automatically removed

2. **Backup Chains**
   - Full backup → Oplog backups form a chain
   - Deleting a full backup will orphan its oplog backups
   - Orphaned oplog backups are automatically cleaned up

## ⚠️ Warning: User and Role Restoration Behavior

Understanding MongoDB user and role restoration behavior is critical for maintaining proper security and access control. Here's a comprehensive guide based on extensive testing with this CLI tool and native MongoDB tools:

1. **Full Backup and Restore**:
   - **Backup Behavior**:
     - Full cluster backups automatically include all user and role definitions from the `admin` database
     - This includes system roles, custom roles, and user credentials across all databases
   - **Restore Options**:
     - choosing to skip the restore of users and roles will behave as expected, the users and roles will not be restored.
     - choosing to restore users and roles will restore all user and role definitions as expected. the user and roles will be restored.

2. **Restoring a single database from a full backup**:
     - **Without users and roles**: Only restores the specified database's data
     - **With users and roles**: Only restores the specified database's data, the users and roles ***will not*** be restored unless the `admin` database is also included in the restore.

3. **Single Database Operations**:
   - **Backup**:
     - Only captures database-specific data
     - Does NOT include user/role definitions
     - Cannot be used for user/role migration
   - **Restore Behavior**:
     - Restores only data, collections, and indexes
     - Users and roles can not be restored from a single database backup


## Contribution

Contributions are welcome! Feel free to submit a pull request or report issues in the [issues section](https://github.com/yourusername/mongodb-backup-cli/issues).
