package models

type PreviousOplogRunInfo struct {
	OplogTakenFrom string `json:"oplog_taken_from"`
	OplogTakenTo   string `json:"oplog_taken_to"`
}
