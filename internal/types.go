package internal

type Config struct {
	DbHost    string `mapstructure:"DB_HOST" validate:"required"`
	DbName    string `mapstructure:"DB_NAME" validate:"required"`
	DbUser    string `mapstructure:"DB_USER" validate:"required"`
	DbPass    string `mapstructure:"DB_PASSWORD" validate:"required"`
	AwsID     string `mapstructure:"AWS_ID" validate:"required"`
	AwsSecret string `mapstructure:"AWS_SECRET" validate:"required"`
	AwsBucket string `mapstructure:"AWS_BUCKET" validate:"required"`
}
