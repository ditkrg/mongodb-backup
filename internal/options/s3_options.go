package options

type S3Options struct {
	EndPoint    string `env:"ENDPOINT,required"`
	AccessKey   string `env:"ACCESS_KEY,required"`
	SecretKey   string `env:"SECRET_ACCESS_KEY,required"`
	Bucket      string `env:"BUCKET,required"`
	Prefix      string `env:"PREFIX"`
	KeepRecentN int    `env:"KEEP_RECENT_N,default=5"`
}
