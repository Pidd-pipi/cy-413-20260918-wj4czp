package config

import "github.com/caarlos0/env/v11"

type Config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"host=localhost user=mindgarden_user password=mindgarden_pwd dbname=mindgarden_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"change_me_to_a_long_random_string"`
	JWTIssuer   string `env:"JWT_ISSUER" envDefault:"mindgarden"`
	CORSOrigin  string `env:"CORS_ORIGIN" envDefault:"http://localhost:18413"`
}

func Load() (Config, error) { return env.ParseAs[Config]() }
