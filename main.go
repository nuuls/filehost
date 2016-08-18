package filehost

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

var log = initLogger(logging.DEBUG)
var cfg = loadConfig()

type config struct {
	Key       string `json:"key"`
	BaseURL   string `json:"base_url"`
	UrlLength int    `json:"url_length"`
}

func Init(m *mux.Router) {
	log.Debug("starting")
	m.HandleFunc("/upload", upload)
	m.HandleFunc(`/{id:[\w\.]+}`, serveFile)
}

func loadConfig() *config {
	file, err := ioutil.ReadFile("filehost.json")
	if err != nil {
		log.Fatal("no config file found")
	}
	var c config
	err = json.Unmarshal(file, &c)
	if err != nil {
		log.Fatal(err)
	}
	if len(c.Key) < 3 {
		log.Warning("the key is very short and may be cracked easily")
	}
	if c.UrlLength < 1 {
		c.UrlLength = 5
	} else if c.UrlLength < 3 {
		log.Warning("the url length is very short and direct links can be guessed easily")
	}
	return &c
}

func initLogger(level logging.Level) *logging.Logger {
	var logger *logging.Logger
	logger = logging.MustGetLogger("filehost")
	logging.SetLevel(level, "filehost")
	backend := logging.NewLogBackend(os.Stdout, "", 0)

	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05.000} %{shortfile:-15s} %{level:.4s}%{color:reset} %{message}`,
	)
	logging.SetFormatter(format)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(level, "filehost")
	logging.SetBackend(backendLeveled)
	return logger
}
