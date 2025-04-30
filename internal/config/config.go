package config

import (
	"errors"

	gEnv "github.com/Netflix/go-env"
)

var (
	ErrInvalidEnv = errors.New("invalid env")
)

type Environment struct {
	Constants
	RunMode string `env:"RUN_MODE,required=true"`
	Port    int    `env:"PORT,default=8080"`

	CorsAllowOrigin string `env:"CORS_ALLOW_ORIGIN,default=*"`

	MysqlHost        string `env:"MYSQL_HOST,required=true"`
	MysqlPort        int    `env:"MYSQL_PORT,required=true"`
	MysqlUser        string `env:"MYSQL_USER,required=true"`
	MysqlPassword    string `env:"MYSQL_PASSWORD,required=true"`
	MysqlDatabase    string `env:"MYSQL_DATABASE,required=true"`
	MysqlSSLMode     string `env:"MYSQL_SSL_MODE,default=disable"`
	MysqlMigrateMode string `env:"MYSQL_MIGRATE_MODE,default=auto"`

	AccessTokenSecretKey     string `env:"ACCESS_TOKEN_SECRET_KEY,required=true"`
	AccessTokenExpireMinutes int    `env:"ACCESS_TOKEN_EXPIRE_MINUTES,default=60"`

	RefreshTokenSecretKey     string `env:"REFRESH_TOKEN_SECRET_KEY,required=true"`
	RefreshTokenExpireMinutes int    `env:"REFRESH_TOKEN_EXPIRE_MINUTES,default=60"`
}

func Load() (*Environment, error) {
	var env Environment

	_, err := gEnv.UnmarshalFromEnviron(&env)

	if err != nil {
		return nil, err
	}

	return &env, nil
}
