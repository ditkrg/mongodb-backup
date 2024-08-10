package flags

type Flags struct {
	S3    S3Flags           `embed:"" prefix:"s3-" envprefix:"S3__" group:"Common S3 Flags:" `
	Mongo MongoRestoreFlags `embed:"" prefix:"mongo-" envprefix:"MONGO_RESTORE__" group:"Common Mongo Restore Flags:"`
}
