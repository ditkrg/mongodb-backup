
# MongoDB Backup CLI Tool

A Golang-powered CLI tool for managing MongoDB backups with seamless integration to S3. The tool supports various features including taking backups, restoring from S3, listing backups, and handling oplog backups.

## Features

- **Backup MongoDB**: Create a complete backup of your MongoDB database and upload it to S3.
- **Restore from S3**: Download a backup from S3 and restore it locally or to a MongoDB instance.
- **List Backups**: View all available backups stored in S3, filtered by type or database.
- **Oplog Backup**: Perform incremental backups using MongoDB's oplog for real-time changes.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/ditkrg/mongodb-backup.git
   cd mongodb-backup
   ```

2. Build the binary:
   ```bash
   go build -o mongodb-backup
   ```

3. Add the binary to your PATH (optional):
   ```bash
   sudo mv mongodb-backup /usr/local/bin
   ```

## Commands and Flags

This CLI tool supports three primary commands: `list`, `dump`, and `restore`, each with a set of configurable flags to suit various MongoDB backup and restoration needs.

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

**Verbosity Options**:
- `--verbosity-level=1 ($VERBOSITY__LEVEL)`: Log verbosity level (1-3, higher is more verbose).
- `--verbosity-quiet ($VERBOSITY__QUIET)`: Suppress all log output.

---

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

---

#### 3. **`restore`**: Restore a database
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

---

## 📋 Notes

### Dump

To successfully perform a `dump` operation using this CLI tool, ensure the MongoDB user has the necessary permissions. The following roles and privileges are recommended:

- The user must have the `backup` role or equivalent permissions on all databases, including the `admin` database. This allows the user to:
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

To successfully perform a `restore` operation using this CLI tool, ensure the MongoDB user has the necessary permissions. The following roles and privileges are recommended:

- The user must have `restore` privileges on all databases, including the `admin` database, to:
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

### Oplog Dump

The oplog user must have the same permissions as the backup user above, besides that you can only perform oplog dump if there is a full backup in the S3 bucket.


### Oplog Restore

- To perform an oplog restore, the user must have a specific role with the following privileges:

```json
{
  "actions": ["anyAction"],
  "resource": { "anyResource": true }
}
```
You can create this user by running the following commands in the MongoDB shell:

```bash
use admin

db.createRole({
  role: "myroot",
  privileges: [{ actions: ["anyAction"], resource: { anyResource: true } }],
  roles: []
})

db.adminCommand({
  createUser: "<Username>",
  pwd: "<Password>",
  roles: [
    { role: "root", db: "admin" },
    { role: "myroot", db: "admin" }
  ]
})
```

- oplog restore will automatically run when trying to restore a backup, the Oplog restore will only work if you do a **full backup** restore.

---

## Examples

### Full Backup Workflow
1. Take a backup:
   ```bash
   mongodb-backup dump --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --connection-string mongodb://localhost:27017
   ```

2. List backups (view the full backups):
   ```bash
   mongodb-backup list --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --full-backups
   ```

3. Restore a backup:
   ```bash
   mongodb-backup restore --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --connection-string mongodb://localhost:27017
   ```

### Oplog Restore

1. Take an oplog backup (there has to be a full backup in the S3 bucket):
   ```bash
   mongodb-backup dump --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --connection-string mongodb://localhost:27017 --oplog
   ```

2. List the backups (view the oplog backups):
   ```bash
   mongodb-backup list --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --oplog
   ```

3. Restore the oplog: after restoring the full backup, the oplog restore will automatically run.
   ```bash
   mongodb-backup restore --s3-endpoint https://mys3.com --s3-access-key MYKEY --s3-secret-key MYSECRET --s3-bucket my-backups --connection-string mongodb://localhost:27017
   ```
---

### ⚠️ Warning: User and Role Restoration Behavior

When working with MongoDB backups and restores, specific behaviors related to users and roles must be noted. Below are the findings based on multiple test scenarios conducted using both this CLI tool and `mongodump`/`mongorestore`:

1. **Full Backup and Restore**:
   - Performing a full backup of the cluster includes both the user and role definitions (typically stored in the `admin` database) along with the data.
   - **Full Restore**:
     - If you choose not to restore users and roles, the restoration excludes them as expected.
     - If you choose to restore users and roles, they are restored correctly.

2. **Partial Restore from a Full Backup**:
   - When restoring a single database from a full backup:
     - **Without Users and Roles**: The restore works as expected, with only the selected database being restored without users and roles.
     - **With Users and Roles**: Only the data for the selected database is restored; users and roles are not restored unless the `admin` database is also included in the restore.
       - **Important**: Including the `admin` database in the restore will restore all users and roles across the cluster, including users from other databases, which may not be desirable.

3. **Single Database Backup and Restore**:
   - Taking a backup of a single database does not include the user and role definitions from the `admin` database.
   - **Restoration Behavior**:
     - A single database restore will only include the database data and no users or roles.

#### Recommendations:
- For scenarios where user and role definitions are critical, always include the `admin` database in your backup and restore operations.
- When restoring a single database with users and roles, ensure you have a clear strategy to avoid unintentionally restoring users from other databases.
- If users and roles for a specific database are needed, consider creating and managing them separately to avoid relying solely on the database restore process.

---

## Contribution

Contributions are welcome! Feel free to submit a pull request or report issues in the [issues section](https://github.com/yourusername/mongodb-backup-cli/issues).
