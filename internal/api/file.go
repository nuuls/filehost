package api

import (
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

func (a *API) serveFile(w http.ResponseWriter, r *http.Request) {
	l := a.log
	name := chi.URLParam(r, "filename")
	name = filepath.Base(name)
	l = l.WithField("file", name)
	// if ratelimited(r.RemoteAddr) {
	// 	l.Warning("ratelimited")
	// 	http.Error(w, "Rate Limit Exceeded", 429)
	// 	return
	// }
	file, err := a.files.Get(name)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}
	defer file.Close()
	spl := strings.Split(name, ".")
	extension := ""
	if len(spl) > 1 {
		extension = spl[len(spl)-1]
	}
	mimeType := ""
	if mimeType == "" {
		mimeType = mime.TypeByExtension("." + extension)
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
	}
	if mimeType == MimeTypeOctetStream {
		switch extension {
		case "mp3":
			mimeType = "audio/mpeg"
		case "wav":
			mimeType = "audio/wav"
		default:
			mimeType = "text/plain"
		}
	}
	if strings.HasPrefix(mimeType, "text") {
		mimeType += "; charset=utf-8"
	}
	if dl, _ := strconv.ParseBool(r.URL.Query().Get("download")); dl {
		mimeType = MimeTypeOctetStream
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", mimeType)

	http.ServeContent(w, r, "", time.Time{}, file)

	err = a.db.IncUploadViews(name)
	if err != nil {
		l.WithError(err).Error("Failed to increment file views")
		return
	}
}
