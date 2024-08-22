package flags

type S3Flags struct {
	EndPoint  string `required:"" name:"s3-endpoint" help:"S3 endpoint"  env:"S3__ENDPOINT"`
	AccessKey string `required:"" name:"s3-access-key" help:"S3 access key"  env:"S3__ACCESS_KEY"`
	SecretKey string `required:"" name:"s3-secret-key" help:"S3 secret access key" env:"S3__SECRET_ACCESS_KEY"`
	Bucket    string `required:"" name:"s3-bucket" help:"S3 bucket" env:"S3__BUCKET"`
	Prefix    string `name:"s3-prefix" help:"S3 Prefix" env:"S3__PREFIX"`
}
