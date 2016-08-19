package filehost

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/op/go-logging"
)

var log = initLogger(logging.DEBUG)
var cfg = loadConfig()
var defaultAllowed = []string{"png", "jpg", "jpeg", "gif", "gifv", "mp3", "mp4", "txt"}
var uploadTempl = template.Must(template.ParseFiles("upload.html"))

type config struct {
	Key        string   `json:"key"`
	BaseURL    string   `json:"base_url"`
	UrlLength  int      `json:"url_length"`
	Blocked    []string `json:"blocked"`
	Allowed    []string `json:"allowed"`
	UploadPage bool     `json:"upload_page"`
}

func (c *config) isBlocked(s string) bool {
	allowed := c.Allowed
	if len(c.Allowed) == 0 {
		allowed = defaultAllowed
	}
	var b bool = true
	for _, e := range allowed {
		if e == s {
			return false
		}
	}
	for _, e := range c.Blocked {
		if e == s {
			b = true
		}
	}
	return b
}

func Init(m *mux.Router) {
	log.Debug("starting")
	if cfg.UploadPage {
		m.HandleFunc("/", index)
	}
	m.HandleFunc("/upload", upload)
	m.HandleFunc(`/{id:[\w\.]+}`, serveFile)
}

func index(w http.ResponseWriter, r *http.Request) {
	uploadTempl.Execute(w, cfg)
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
