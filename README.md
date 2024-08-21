# Project Name

## Introduction
This project is a Go CLI Tool that is used for backing up and restoring MongoDB databases. To test or to view the possible commands and their flags, you can run the following command:

```bash
go run main.go --help
```

you can check each commands flags/environment variables by running the following command:

```bash
go run main.go <command> --help
```

## global Log Level
You can set the log level for the whole application by setting the `LOG_LEVEL` environment variable, the possible values are `Debug`, `Info`, `Warn`, `Error`, `Fatal`, `Panic`, `Disable`, `Trace`. The default value is `Error`.

mongorestore  --gzip --oplogReplay --dir=./backups --uri="mongodb://root:root@localhost:27017/?authSource=admin&tls=false&replicaSet=replicaset"

db.createRole({role: "myroot", privileges: [{actions: [ "anyAction" ],resource: { anyResource: true }}],roles: []})
db.adminCommand({createUser: "root",pwd: "root",roles: [{ role: "root", db: "admin" },{ role: "myroot", db: "admin" }]})
db.createUser({user:"root",pwd: "root",roles: [{ role: "root", db: "admin" }]})

db.updateUser("root2",{roles : [{ role: "myroot", db: "admin" },{ role: "root", db: "admin" }]})
