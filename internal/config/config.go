package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Addr                     string       `default:":7417"`
	LogLevel                 logrus.Level `default:"debug"`
	PostgresDSN              string       `default:"host=localhost user=postgres password=postgrespw dbname=postgres port=49153 sslmode=disable"`
	StorageBucketEndpoint    string       `envconfig:"STORAGE_BUCKET_ENDPOINT"`
	StorageBucketName        string       `envconfig:"STORAGE_BUCKET_NAME"`
	StorageBucketAccessKeyID string       `envconfig:"STORAGE_BUCKET_ACCESS_KEY_ID"`
	StorageBucketSecretKey   string       `envconfig:"STORAGE_BUCKET_SECRET_KEY"`
	FallbackFilePath         string       `default:"./files"`
	DefaultDomainID          uint         `default:"1"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file", err)
	}
	cfg := &Config{}
	err = envconfig.Process("", cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
	return cfg
}
