package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nuuls/filehost"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := loadConfig()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "02-01 15:04:05.000",
	})
	conf := &filehost.Config{
		AllowedMimeTypes: cfg.AllowedMimeTypes,
		Authenticate: func(r *http.Request) bool {
			if cfg.Password == "" {
				return true
			}
			password := r.URL.Query().Get("password")
			if password == "" {
				password = r.Header.Get("password")
			}
			if password == cfg.Password {
				return true
			}
			return false
		},
		Logger: logrus.StandardLogger(),
		NewFileName: func() string {
			return filehost.RandString(cfg.NameLength)
		},
		BasePath: cfg.BasePath,
		BaseURL:  cfg.BaseURL,
		AllowFileName: func(r *http.Request) bool {
			if cfg.MasterPassword == "" {
				return false
			}
			password := r.URL.Query().Get("password")
			if password == "" {
				password = r.Header.Get("password")
			}
			if password == cfg.MasterPassword {
				return true
			}
			return false
		},
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/*", filehost.New(conf))

	log.Fatal(http.ListenAndServe(":7494", r))
}

type config struct {
	Host             string
	AllowedMimeTypes []string
	BasePath         string
	Password         string
	MasterPassword   string
	BaseURL          string // https://i.nuuls.com/
	NameLength       int
}

func loadConfig() *config {
	file, err := ioutil.ReadFile("filehost.json")
	if err != nil {
		log.Fatal(err)
	}
	var c config
	err = json.Unmarshal(file, &c)
	if err != nil {
		log.Fatal(err)
	}
	return &c
}
