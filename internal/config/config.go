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

	// Server Config
	RunMode string `env:"RUN_MODE,required=true"`
	Port    int    `env:"PORT,default=8080"`

	// Cors Config
	CorsAllowOrigins string `env:"CORS_ALLOW_ORIGIN,default=*"`

	// Mysql DB Config
	MysqlHost            string `env:"MYSQL_HOST,required=true"`
	MysqlPort            int    `env:"MYSQL_PORT,required=true"`
	MysqlUser            string `env:"MYSQL_USER,required=true"`
	MysqlPassword        string `env:"MYSQL_PASSWORD,required=true"`
	MysqlDatabase        string `env:"MYSQL_DATABASE,required=true"`
	MysqlSSLMode         string `env:"MYSQL_SSL_MODE,default=disable"`
	MysqlMigrateMode     string `env:"MYSQL_MIGRATE_MODE,default=auto"`
	MysqlMaxOpenConns    int    `env:"MYSQL_MAX_OPEN_CONNS,default=100"`
	MysqlMaxIdleConns    int    `env:"MYSQL_MAX_IDLE_CONNS,default=10"`
	MysqlConnMaxLifetime int    `env:"MYSQL_CONN_MAX_LIFETIME,default=60"`
	MysqlConnMaxIdleTime int    `env:"MYSQL_CONN_MAX_IDLE_TIME,default=60"`

	// JWT Config
	AccessTokenSecretKey     string `env:"ACCESS_TOKEN_SECRET_KEY,required=true"`
	AccessTokenExpireMinutes int    `env:"ACCESS_TOKEN_EXPIRE_MINUTES,default=60"`

	RefreshTokenSecretKey     string `env:"REFRESH_TOKEN_SECRET_KEY,required=true"`
	RefreshTokenExpireMinutes int    `env:"REFRESH_TOKEN_EXPIRE_MINUTES,default=60"`

	// FE Config
	FrontendUri  string `env:"FRONTEND_URI,required=true"`
	FrontendPort int    `env:"FRONTEND_PORT,required=true"`
}

func Load() (*Environment, error) {
	var env Environment

	_, err := gEnv.UnmarshalFromEnviron(&env)

	if err != nil {
		return nil, err
	}

	return &env, nil
}
