// Package env loads typed configuration structs from environment variables.
//
// Struct fields use the tag format from github.com/caarlos0/env/v11:
//
//	type Config struct {
//		Addr    string        `env:"ADDR" envDefault:":8080"`
//		Timeout time.Duration `env:"TIMEOUT" envDefault:"5s"`
//		DSN     string        `env:"DATABASE_URL,required,notEmpty"`
//	}
//
// The same tags are compatible with github.com/g4s8/envdoc, so application
// config packages can generate environment variable documentation from the
// struct comments they already maintain.
package env
