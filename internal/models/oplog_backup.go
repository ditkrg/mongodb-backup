package models

import "time"

type PitrBackup struct {
	Key                      string
	FileName                 string
	FileNameWithoutExtension string
	FromString               string
	FromTime                 time.Time
	ToString                 string
	ToTime                   time.Time
}
