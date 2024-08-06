package helpers

const OplogQuery = "{ \"ts\": { \"$gt\": { \"$timestamp\": { \"t\": %d, \"i\": 1 } } } }"
const ConfigFileName = "oplog_config.json"
