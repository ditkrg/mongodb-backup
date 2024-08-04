package options

type Options struct {
	MongoDB MongoDBOptions `env:",prefix=MONGODB__"`
	S3      S3Options      `env:",prefix=S3__"`
}

func (o *Options) Validate() error {
	if err := o.MongoDB.Validate(); err != nil {
		return err
	}

	return nil
}
