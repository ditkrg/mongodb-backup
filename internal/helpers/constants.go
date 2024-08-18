package helpers

const (
	OplogQuery     = "{ \"wall\": { \"$gt\": {\"$date\": \"%s\"}, \"$lte\": {\"$date\": \"%s\"} } }"
	TimeFormat     = "2006-01-02T15:04:05.000-07:00"
	ConfigFileName = "oplog_config.json"
)
