package flags

type S3Flags struct {
	EndPoint  string `required:"" help:"S3 endpoint"  env:"ENDPOINT"`
	AccessKey string `required:"" help:"S3 access key"  env:"ACCESS_KEY"`
	SecretKey string `required:"" help:"S3 secret access key" env:"SECRET_ACCESS_KEY"`
	Bucket    string `required:"" help:"S3 bucket" env:"BUCKET"`
	Prefix    string `help:"S3 Prefix" env:"PREFIX"`
}
