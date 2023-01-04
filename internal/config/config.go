package config

import "github.com/sirupsen/logrus"

type Config struct {
	Addr             string
	LogLevel         logrus.Level
	PostgresURI      string
	StorageBucketURI string
}
