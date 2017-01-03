package filehost

import (
	"context"
	"io"
	"net/http"
	"time"

	"os"

	"path/filepath"

	"github.com/pressly/chi"
	"github.com/sirupsen/logrus"
)

var log logrus.FieldLogger
var cfg *Config

type Config struct {
	AllowedMimeTypes []string
	BasePath         string
	NewFileName      func() string
	SaveFileInfo     func(*FileInfo)
	GetFileInfo      func(string) *FileInfo
	Authenticate     func(*http.Request) bool
	AllowFileName    func(*http.Request) bool
	Logger           logrus.FieldLogger
}

type FileInfo struct {
	Name     string // filename without extension
	Path     string
	MimeType string
	Uploader interface{} // information about the person who uploaded it
	Time     time.Time
	Expire   time.Duration
}

func New(conf *Config) http.Handler {
	cfg = conf
	log = cfg.Logger
	err := os.MkdirAll(cfg.BasePath, 644)
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), "logger",
				log.WithFields(logrus.Fields{
					"ip":         r.RemoteAddr,
					"user-agent": r.UserAgent(),
				})))
			log.Debug("xd")
			next.ServeHTTP(w, r)
		})
	})
	r.Post("/upload", upload)
	return r
}

func upload(w http.ResponseWriter, r *http.Request) {
	l := r.Context().Value("logger").(logrus.FieldLogger)
	l.Debug("NaM")
	defer r.Body.Close()
	if !cfg.Authenticate(r) {
		http.Error(w, "Not Authenticated", http.StatusUnauthorized)
		return
	}
	err := r.ParseMultipartForm(1024 * 1024 * 64)
	if err != nil {
		l.WithError(err).Error("cannot read multi part form")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if r.MultipartForm == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File
	l.Debug(files)
	if len(files) < 1 {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	for _, headers := range files {
		for _, h := range headers {
			file, err := h.Open()
			if err != nil {
				l.Error(err)
			}
			name := h.Filename
			if !cfg.AllowFileName(r) {
				name = cfg.NewFileName()
			}
			l = l.WithField("file", name)
			l.Info("uploading...")
			// TODO: check mime type and append file extension if needed
			dstPath := filepath.Join(cfg.BasePath, name)
			// TODO: check if file exists
			dst, err := os.Create(dstPath)
			if err != nil {
				l.WithError(err).Error("cannot create file")
				http.Error(w, "Internal Server Error", 500)
				return
			}
			_, err = io.Copy(dst, file)
			if err != nil {
				l.WithError(err).Error("failed to save file")
				http.Error(w, "Internal Server Error", 500)
				return
			}
			// TODO: save fileinfo
		}
	}
}
