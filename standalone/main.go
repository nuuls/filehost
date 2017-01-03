package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/nuuls/filehost"

	"github.com/pressly/chi"
	"github.com/pressly/chi/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	//cfg := loadConfig()
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "02-01 15:04:05.000",
	})
	conf := &filehost.Config{
		AllowedMimeTypes: []string{
			"*/*",
		},
		Authenticate: func(r *http.Request) bool {
			return true
		},
		Logger: logrus.New(),
		NewFileName: func() string {
			return strconv.Itoa(rand.Intn(420))
		},
		BasePath: "./files",
		AllowFileName: func(r *http.Request) bool {
			return true
		},
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/*", filehost.New(conf))

	log.Fatal(http.ListenAndServe(":7494", r))
}

type config struct {
	Host string `json:"host"`
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
