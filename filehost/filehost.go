package filehost

import (
	"context"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pressly/chi"
	"github.com/sirupsen/logrus"
)

var log logrus.FieldLogger = logrus.StandardLogger()
var cfg *Config

// Database is some sort of database that stores FileInfo
type Database interface {
	// GetFileInfo returns the FileInfo associated with the given
	// File Name, it returns nil if the file was not found
	GetFileInfo(string) *FileInfo
	SaveFileInfo(*FileInfo)
}

// Config contains the needed information to call New
type Config struct {
	AllowedMimeTypes []string
	BasePath         string
	BaseURL          string
	NewFileName      func() string
	DB               Database
	Authenticate     func(*http.Request) bool
	AllowFileName    func(*http.Request) bool
	Logger           logrus.FieldLogger
}

// FileInfo contains additional information about the File
type FileInfo struct {
	Name     string // filename without extension
	Path     string
	MimeType string
	Uploader interface{} // information about the person who uploaded it
	Time     time.Time
	Expire   time.Duration
	Clicks   int
}

// New initializes a http Handler and returns it
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
			next.ServeHTTP(w, r)
		})
	})
	r.Post("/upload", upload)
	r.Get("/:file", serveFile)
	r.Get("/", http.NotFound)
	return r
}

func upload(w http.ResponseWriter, r *http.Request) {
	l := r.Context().Value("logger").(logrus.FieldLogger)
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
			if !cfg.AllowFileName(r) || name == "" {
				if cfg.NewFileName != nil {
					name = cfg.NewFileName()
				} else {
					name = RandString(5)
				}
			}
			l = l.WithField("file", name)
			l.Info("uploading...")
			l.Debug(r.Header)
			l.Debug(h.Header)
			mimeType := h.Header.Get("Content-Type")
			log.Info(mimeType)
			if !whiteListed(cfg.AllowedMimeTypes, mimeType) {
				l.WithField("mime-type", mimeType).Warning("mime type not allowed")
				http.Error(w, "Forbidden", 403)
				return
			}
			extensions, err := mime.ExtensionsByType(mimeType)
			extension := ".png"
			if len(extensions) < 1 || err != nil {
				l.WithError(err).Warning("no extension found, sniffing mime type...")
				// TODO: sniff for mime type
			} else {
				extension = extensions[0]
			}
			fullName := name + extension
			// TODO: check mime type and append file extension if needed
			dstPath := filepath.Join(cfg.BasePath, fullName)
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
			w.Write([]byte(cfg.BaseURL + fullName))
			l.Info("uploaded to ", cfg.BaseURL+fullName)
			if cfg.DB != nil {
				info := &FileInfo{
					Name: name,
					Path: dstPath,
					Uploader: map[string]interface{}{
						"ip":         r.RemoteAddr,
						"user-agent": r.UserAgent(),
					},
					Time:     time.Now(),
					MimeType: mimeType,
				}
				cfg.DB.SaveFileInfo(info)
			}
		}
	}
}

const octetStream = "application/octet-stream"

func serveFile(w http.ResponseWriter, r *http.Request) {
	l := r.Context().Value("logger").(logrus.FieldLogger)
	name := chi.URLParam(r, "file")
	l = l.WithField("file", name)
	file, err := os.Open(filepath.Join(cfg.BasePath, name))
	if err != nil {
		l.WithError(err).Warning("not found")
		http.Error(w, "404 Not Found", 404)
		return
	}
	spl := strings.Split(name, ".")
	id := spl[0]
	extension := ""
	if len(spl) > 1 {
		extension = spl[len(spl)-1]
	}
	mimeType := ""
	if cfg.DB != nil {
		info := cfg.DB.GetFileInfo(id)
		if info != nil {
			if info.Expire != 0 {
				if time.Since(time.Now().Add(info.Expire)) > info.Expire {
					l.Info("expired")
					http.Error(w, "404 Not Found", 404)
					return
				}
			}
			mimeType = info.MimeType
			info.Clicks++
		}
	}

	if mimeType == "" {
		sniffData := make([]byte, 512)
		n, err := file.Read(sniffData)
		if err != nil {
			l.WithError(err).Error("cannot read from file")
			http.Error(w, "Internal Server Error", 500)
			return
		}
		sniffData = sniffData[:n]
		mimeType = http.DetectContentType(sniffData)
		_, err = file.Seek(0, 0)
		if err != nil {
			l.WithError(err).Error("cannot seek file")
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
	if mimeType == octetStream {
		switch extension {
		case "mp3":
			mimeType = "audio/mpeg"
		case "wav":
			mimeType = "audio/wav"
		default:
			mimeType = "text/plain"
		}
	}
	if r.URL.Query().Get("download") != "" {
		mimeType = octetStream
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", mimeType)

	_, err = io.Copy(w, file)
	if err != nil {
		l.WithError(err).Error("cannot serve file")
	}
}
