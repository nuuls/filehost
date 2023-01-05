package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Addr             string       `default:":7417"`
	LogLevel         logrus.Level `default:"debug"`
	PostgresDSN      string       `default:"host=localhost user=postgres password=postgrespw dbname=postgres port=49153 sslmode=disable"`
	StorageBucketURI string
	DefaultDomainID  uint `default:"1"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
	cfg := &Config{}
	err = envconfig.Process("FH", cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cfg
}
