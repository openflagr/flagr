package config

import (
	"os"

	"github.com/jinzhu/configor"
)

// Config is the whole configuration of the app
var Config = struct {
	DB struct {
		DBDriver        string `required:"true"`
		DBConnectionStr string `env:"DBConnectionStr"`
	}
	CORS struct {
		Enabled bool
	}
}{}

func init() {
	pwd, _ := os.Getwd()
	configor.Load(&Config, pwd+"/pkg/config/config.toml")
}
