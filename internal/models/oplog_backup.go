package models

import "time"

type OplogBackup struct {
	Key                      string
	FileName                 string
	FileNameWithoutExtension string
	FromString               string
	FromTime                 time.Time
	ToString                 string
	ToTime                   time.Time
}
