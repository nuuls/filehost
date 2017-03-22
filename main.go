package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nuuls/filehost/filehost"

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
		UploadPage:      true,
		ExposedPassword: cfg.Password,
		Logger:          logrus.StandardLogger(),
		NewFileName: func() string {
			return filehost.RandString(cfg.NameLength)
		},
		DB:       filehost.NewDB("./db/database.json"),
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
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := r.Header.Get("Cf-Connecting-Ip"); ip != "" {
				r.RemoteAddr = ip
			}
			w.Header().Set("Access-Control-Allow-Origin", "*")

			for _, h := range r.Header["Access-Control-Request-Methods"] {
				w.Header().Add("Access-Control-Allow-Methods", h)
			}

			for _, h := range r.Header["Access-Control-Request-Headers"] {
				w.Header().Add("Access-Control-Allow-Headers", h)
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	r.Use(middleware.Logger)
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
