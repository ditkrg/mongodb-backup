# MongoDB Backup CLI

## Introduction
This project is a Go CLI Tool that is used for backing up and restoring MongoDB databases. To test or to view the possible commands and their flags, you can run the following command:

```bash
<output_file_name> --help
```

you can check each commands flags/environment variables by running the following command:

```bash
<output_file_name> <command> --help
```

## Build CLI Locally
To build the CLI locally, you can run the following command:

```bash
go build -o <output_file_name> main.go
```


## Global Log Level
You can set the log level for the whole application by setting the `LOG_LEVEL` environment variable, the possible values are `Debug`, `Info`, `Warn`, `Error`, `Fatal`, `Panic`, `Disable`, `Trace`. The default value is `Error`.

---

# Database Dump & Restore

## List Database backups
To list the available backups, you can run the following command:

```bash
<output_file_name> list <--full-backups|--database="database name">
```

## Database Dump
To dump a database, you can run the following command:

```bash
<output_file_name> dump
```

make sure to provide the the required configurations either by flags or by environment variables, you can check the available flags/environment variables by running the following command:

```bash
<output_file_name> dump --help
```

## Database Restore
To restore a database, you can run the following command:

```bash
<output_file_name> restore database
```

make sure to provide the the required configurations either by flags or by environment variables, you can check the available flags/environment variables by running the following command:

```bash
<output_file_name> restore database --help
```

---

# Oplog Dump & Restore


## Oplog restore user
To be able to perform the oplog restore, you need to have a user with the following privileges:

```json
{
  "actions": [ "anyAction" ],
  "resource": { "anyResource": true}
}
```

you can create the user by running the following command in the mongoshell:

```bash
db.createRole({role: "myroot", privileges: [{actions: [ "anyAction" ],resource: { anyResource: true }}],roles: []})
db.adminCommand({createUser: "<Username>",pwd: "<Password>",roles: [{ role: "root", db: "admin" },{ role: "myroot", db: "admin" }]})
```

## List Oplog backups
To list the available oplog backups, you can run the following command:

```bash
<output_file_name> list --oplog
```

## Oplog Dump
To take a dump of the oplog, you can run the following command:

```bash
<output_file_name> dump --oplog
```

## Oplog Restore
To restore the oplog, you can run the following command:

```bash
<output_file_name> restore oplog
```
